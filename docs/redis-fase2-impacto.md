# Impacto Fase 2 — Redis para `device_last_positions`

## Resumen ejecutivo

La Fase 2 migró el caché de última posición desde la tabla `device_last_positions` (PostgreSQL) hacia Redis, manteniendo la tabla como respaldo (dual-write). El patrón de lectura es **Redis-first con fallback a Postgres**. No se eliminó la tabla: sigue existiendo como red de seguridad.

---

## Qué cambió

| Archivo | Tipo | Cambio |
|---------|------|--------|
| `providers/redis/device_position_cache.go` | Nuevo | Interfaz + implementación Redis para posiciones |
| `database/entities/position_entity.go` | Modificado | Removido `AfterCreate` SQL hook |
| `modules/position/repository/position_repository.go` | Modificado | Redis en Create, FindLastByIMEI, FindLastPositions, FindLastPositionsByVehicles, FindLastPositionByVehicle |
| `modules/device/repository/dispositivo_repository.go` | Modificado | `GetDevicesWithLastPosition` usa Redis + resolvePositions |
| `modules/position/routes.go` | Modificado | Inyecta `DevicePositionCache` |
| `cmd/main.go` | Modificado | Warm-up goroutine al arrancar |

---

## Flujo de datos actual

### Escritura (cada posición GPS recibida)
```
GPS → position_service.Create()
        └─ db.Create(position)           → tabla positions ✓
        └─ updatePositionCache()
              ├─ Redis SET device:pos:{IMEI}     (fuente principal de lecturas)
              ├─ Redis SADD device:pos:index      (índice para GetAll)
              └─ SQL UPSERT device_last_positions (respaldo dual-write)
```

### Lectura (dashboard, API, admin)
```
Request
  └─ resolvePositions(imeis[])
       ├─ Redis MGET device:pos:{IMEI...}   → hit: retorna directo O(1)
       └─ miss: SQL WHERE imei IN (...)      → fallback device_last_positions
```

### Claves Redis generadas
```
device:pos:{IMEI}    → JSON con posición completa    (TTL: sin expiración)
device:pos:index     → SET con todos los IMEIs        (para GetAll/dashboard)
```

---

## Rutas de lectura afectadas

| Método | Antes | Ahora | Fallback |
|--------|-------|-------|----------|
| `FindLastByIMEI` | `SELECT * FROM device_last_positions WHERE imei=?` | Redis GET | → device_last_positions → positions |
| `FindLastPositions` | `SELECT * FROM device_last_positions` | Redis GetAll (SMEMBERS + MGET) | → device_last_positions |
| `FindLastPositionsByVehicles` | JOIN SQL con device_last_positions | Postgres metadata + Redis MGET | → device_last_positions por IMEIs faltantes |
| `FindLastPositionByVehicle` | JOIN SQL con device_last_positions | Postgres metadata + Redis GET | → device_last_positions |
| `GetDevicesWithLastPosition` | LEFT JOIN SQL con device_last_positions | Postgres metadata + Redis MGET | → device_last_positions |

---

## Problemas identificados

### 🔴 Críticos

#### 1. Race condition en el warm-up
**Archivo**: `cmd/main.go` — goroutine de warm-up  
**Escenario**:
1. Servidor arranca, warm-up empieza en background
2. Dispositivo envía posición nueva → escrita en Redis correctamente
3. Warm-up aún no terminó → lee `device_last_positions` (valor anterior)
4. Warm-up sobreescribe en Redis con el valor **stale**

**Resultado**: La posición más reciente se pierde temporalmente hasta el próximo GPS.

**Fix sugerido**: Al warm-up, usar `SET ... NX` (solo escribir si la clave no existe) para que el warm-up no pise valores recientes:
```go
// En device_position_cache.go: agregar método SetNX
c.redis.Client().SetNX(ctx, devicePosKey(pos.IMEI), string(data), 0)
```

---

#### 2. `SAdd` sin verificar error
**Archivo**: `providers/redis/device_position_cache.go:54`  
```go
c.redis.Client().SAdd(ctx, devicePosIndex, pos.IMEI)  // error ignorado
```
Si `SAdd` falla, la clave `device:pos:{IMEI}` existe en Redis pero el IMEI **no está en el índice**. `GetAll()` usará `SMEMBERS` y nunca verá ese dispositivo.

**Impacto en endpoints**: `GET /api/positions/latest` puede devolver lista incompleta.

**Fix**:
```go
if err := c.redis.Client().SAdd(ctx, devicePosIndex, pos.IMEI).Err(); err != nil {
    log.Printf("[pos-cache] error al agregar IMEI al índice: %v", err)
}
```

---

### 🟡 Altos

#### 3. Dispositivos sin posición silenciosamente eliminados de resultados
**Archivo**: `modules/position/repository/position_repository.go:333`  
```go
if !ok {
    continue  // vehiculo sin posición → omitido del resultado
}
```
El JOIN original era `JOIN` (inner), así que este comportamiento **es correcto** para `FindLastPositionsByVehicles`. Pero para `GetDevicesWithLastPosition` el original era `LEFT JOIN` — los dispositivos sin posición deben aparecer con campos nulos. En el código nuevo sí se manejan con punteros nil (línea 238 de dispositivo_repository.go), lo cual está correcto.

**Verificar** que el frontend maneja `null` en los campos de posición sin romper.

---

#### 4. `resolvePositions` retorna resultados parciales sin indicar error
**Archivo**: `modules/position/repository/position_repository.go:388`  
Si Redis falla Y la query de fallback a Postgres también falla, la función retorna un mapa vacío o parcial **sin error**. El caller asume que el resultado está completo.

**Impacto**: Dashboard puede mostrar lista vacía sin indicación de problema.

---

### 🟠 Medios

#### 5. `FindLastPositions` sin orden
**Archivo**: `modules/position/repository/position_repository.go:197`  
El path Redis (`GetAll`) retorna posiciones en orden arbitrario (Redis SET no tiene orden). El path fallback tampoco tiene `ORDER BY`. El SQL anterior seguramente tenía un orden consistente.

**Impacto**: El endpoint `GET /api/positions/latest` puede devolver orden diferente en cada llamada.

---

#### 6. IMEI vacío crea clave malformada en Redis
**Archivo**: `modules/position/repository/position_repository.go:80`  
Si `p.Imei == ""`, se crea la clave `device:pos:` sin sufijo.

**Fix**:
```go
if p.Imei == "" {
    log.Printf("[pos-cache] posición sin IMEI, saltando caché")
    return
}
```

---

#### 7. `Get()` no distingue "clave inexistente" de "Redis caído"
**Archivo**: `providers/redis/device_position_cache.go:58`  
Ambos casos retornan `(CachedPosition{}, false, nil)`. Cuando Redis está caído, se van innecesariamente al fallback de Postgres para cada request, pudiendo sobrecargar la DB.

---

#### 8. Warm-up sin timing ni retry
El warm-up no reporta cuánto tardó ni reintenta si falla la lectura de Postgres. Si la tabla está vacía o la DB no responde al arrancar, Redis arranca completamente frío sin aviso visible.

---

#### 9. Dispositivos desinstalados siguen en el caché
`device_last_positions` no tiene campo `removed_at`. Cuando un dispositivo se desinstala (`device_installations.removed_at IS NOT NULL`), su posición permanece en Redis indefinidamente. Las queries que filtran por `di.removed_at IS NULL` en el JOIN de metadata de vehículos no exponen este dato, pero `FindLastByIMEI` lo seguiría devolviendo.

---

## Diferencias de comportamiento vs. SQL original

| Comportamiento | Antes (SQL) | Ahora (Redis) |
|----------------|-------------|---------------|
| Orden de `FindLastPositions` | Definido por ORDER BY | Arbitrario (SET Redis) |
| Dispositivo sin posición en `FindLastPositionsByVehicles` | Excluido (INNER JOIN) | Excluido ✓ |
| Dispositivo sin posición en `GetDevicesWithLastPosition` | NULL en campos (LEFT JOIN) | NULL en punteros ✓ |
| Error de lectura | SQL error propagado | Silenciado, fallback a tabla |
| Transaccionalidad del caché | En la misma TX de `positions` | Fuera de TX (eventual) |
| Posición de dispositivo removido | Excluida por el JOIN | Visible vía `FindLastByIMEI` |

---

## Implicaciones de memoria y performance

### Claves Redis
- Por dispositivo: ~120 bytes (clave + JSON posición)
- Índice `device:pos:index`: ~120 bytes por IMEI
- **1.000 dispositivos** ≈ 240 KB en Redis (insignificante)
- **100.000 dispositivos** ≈ 24 MB en Redis (acceptable)

### `GetAll()` — flujo del dashboard
1. `SMEMBERS device:pos:index` → O(N) — retorna todos los IMEIs
2. `MGET device:pos:imei1 device:pos:imei2 ...` → O(N) — un solo round-trip

Antes: query SQL completa a tabla. Ahora: dos comandos Redis, ambos O(N) pero **sin I/O a disco**.

### Dos queries Postgres en `FindLastPositionsByVehicles`
El método ahora hace **dos queries separadas** a Postgres (metadata de instalación + fallback por IMEIs faltantes) vs. el JOIN original (una sola query). Si Redis tiene alta tasa de acierto, la segunda query nunca ocurre. Si Redis está frío (cold start), puede haber más carga a Postgres que antes.

---

## Estado de la tabla `device_last_positions`

La tabla **sigue existiendo** y se escribe en cada posición (dual-write). Se puede eliminar en una **Fase 2d** cuando:

1. El warm-up use `SetNX` (para evitar la race condition del punto 1)
2. Se valide en producción que Redis tiene >99% hit rate
3. Se confirme que el fallback nunca se activa durante operación normal

Para eliminarla: remover el bloque SQL UPSERT en `updatePositionCache`, remover `resolvePositions` fallback a Postgres, y eliminar la migración/entidad `DeviceLastPosition`.

---

## Fixes prioritarios recomendados

```
Prioridad 1 (esta semana):
  ├─ Corregir SAdd error check en device_position_cache.go:54
  └─ Corregir warm-up para usar SetNX en lugar de Set

Prioridad 2 (próximo sprint):
  ├─ Agregar validación de IMEI vacío antes de escribir caché
  └─ Loggear cuando resolvePositions retorna resultados parciales por error

Prioridad 3 (cuando Redis sea estable):
  └─ Eliminar dual-write a device_last_positions (Fase 2d)
```
