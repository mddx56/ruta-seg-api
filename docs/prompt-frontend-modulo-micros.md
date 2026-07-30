# Prompt para el equipo/IA de Frontend — Panel Web Admin del Módulo "Micros"

> Este documento es el **contrato técnico completo** del backend (`ruta-seg-api`) para el módulo de microbuses, con foco exclusivo en el **panel web de administración**. `ruta-seg-api` es solo backend: no hay código de UI aquí, todo lo que sigue son endpoints REST y eventos WebSocket ya implementados y probados en vivo (ver `docs/plan-modulo-micros.md` para el detalle de negocio y las fases).

## 0. Contexto de negocio (resumen)

Se está agregando transporte público ("micros"/buses) al sistema, que hoy solo trackea motos. Piloto: **Línea 18** (anillo cerrado). Este documento cubre el backend que consume el **panel web de administración** (rol `admin`): mapa con todos los micros (número en el pin), conteo de vueltas, cobro por vuelta, multas por infracciones (adelantamiento, parada prolongada, tiempo de vuelta fuera de lo permitido, exceso de velocidad), y CRUD de rutas/tarifas/asignaciones.

Regla de negocio clave: el panel Admin debe tener un filtro **"EN RUTA"**. Sin el filtro se ven todos los micros asignados; con el filtro activo, solo los que tienen posición reciente sobre el trazado de su ruta — eso es exactamente lo que devuelve `GET /api/routes/:id/live`.

## 1. Convenciones generales del backend

**Envelope de toda respuesta HTTP**:
```json
{ "status": true, "message": "Éxito", "error": null, "data": { }, "meta": null }
```
`status=false` en error, `error` trae el detalle, `data`/`error` se omiten si no aplican.

**Autenticación**: JWT como header `Authorization: Bearer <token>` en todos los endpoints salvo los marcados explícitamente como públicos. Errores típicos:
- Sin header o sin `Bearer `: `401`, `error: "token no encontrado"` / `"token no valido"`.
- Token expirado/inválido: `401`, `error: "token no valido"`.
- Ruta admin-only sin rol admin: `403`, `error: "Acceso denegado: se requiere rol de administrador"`.

**IDs**: todos UUID v4, salvo `Position.ID` (uint64, autoincremental).

**Fechas**: ISO 8601 (`RFC3339`), UTC.

---

## 2. Referencia completa de endpoints REST

### 2.1 Rutas — `modules/route` (base `/api/routes`)

| Método | Path | Auth |
|---|---|---|
| POST | `/api/routes` | admin |
| GET | `/api/routes` | público |
| GET | `/api/routes/:id` | público |
| GET | `/api/routes/:id/live` | público |
| GET | `/api/routes/:id/eta?lat=&lon=` | público |
| PUT | `/api/routes/:id` | admin |
| PATCH | `/api/routes/:id/status` | admin |

`POST`/`PUT` body:
```json
{
  "name": "Línea 25", "description": "...", "map_color": "#FF0000",
  "checkpoint_geofence_id": "uuid",
  "allowed_lap_duration_seconds": 3600, "max_speed_kmh": 60, "max_stop_duration_seconds": 300,
  "geometry": "{...GeoJSON serializado...}",
  "stops": [ { "name": "Terminal", "latitude": -17.79, "longitude": -63.17, "sequence": 1 } ]
}
```
En `PUT`, si `stops` viene presente reemplaza toda la lista existente. Todos los campos son opcionales salvo `name` en el `POST`.

`RouteResponse` (forma de `data` en create/update/get):
```json
{
  "id": "uuid", "name": "...", "description": "...", "map_color": "#FF0000",
  "checkpoint_geofence_id": "uuid",
  "allowed_lap_duration_seconds": 3600, "max_speed_kmh": 60, "max_stop_duration_seconds": 300,
  "geometry": "{...}", "active": true,
  "stops": [ { "id": "uuid", "name": "Terminal", "latitude": -17.79, "longitude": -63.17, "sequence": 1 } ],
  "created_by_id": "uuid", "created_at": "2026-07-07T00:00:00Z", "status": true
}
```

`GET /api/routes/:id/live` → `data: []LiveVehicleResponse`, solo micros con posición de menos de **2 minutos**:
```json
[{ "vehicle_id": "uuid", "pin_number": "4", "latitude": -17.79, "longitude": -63.17,
   "speed_kmh": 32, "last_update_at": "...", "seconds_since_update": 15,
   "lap_number": 3, "lap_status": "IN_PROGRESS" }]
```
(`lap_number`/`lap_status` se omiten si no hay vuelta abierta). **Este es el endpoint para el filtro "EN RUTA".**

`GET /api/routes/:id/eta?lat=<float>&lon=<float>` → `data: []EtaResponse`, ordenado por ETA ascendente. No es prioritario para el panel Admin (pensado originalmente para consumo público), pero queda documentado por si se reutiliza en algún reporte:
```json
[{ "vehicle_id": "uuid", "pin_number": "4", "distance_meters": 850.3, "eta_seconds": 127 }]
```

### 2.2 Tarifas por vuelta — `modules/route_fare` (base `/api/route-fares`)

| Método | Path | Auth |
|---|---|---|
| POST | `/api/route-fares` | admin |
| GET | `/api/route-fares` | user |
| GET | `/api/route-fares/:id` | user |
| PUT | `/api/route-fares/:id` | admin |
| PATCH | `/api/route-fares/:id/status` | admin |

Body: `{ "route_id": "uuid", "amount_per_lap": 5, "effective_from": "2026-07-01T00:00:00Z" }` (`effective_from` opcional). Respuesta con los mismos campos + `id`, `created_at`, `status`.

### 2.3 Asignación micro↔ruta — `modules/vehicle_route` (base `/api/vehicle-routes`)

| Método | Path | Auth |
|---|---|---|
| POST | `/api/vehicle-routes` | admin |
| GET | `/api/vehicle-routes` | user |
| GET | `/api/vehicle-routes/:id` | user |
| PUT | `/api/vehicle-routes/:id` | admin |
| PATCH | `/api/vehicle-routes/:id/status` | admin |

Body: `{ "vehicle_id": "uuid", "route_id": "uuid", "pin_number": "4", "assigned_at": "..." }` (`assigned_at` opcional). Restricción: `(route_id, pin_number)` único — no puede repetirse el pin en la misma ruta. Respuesta agrega `id`, `active`, `created_at`, `status`.

### 2.4 Vueltas (solo lectura) — `modules/lap` (base `/api/laps`)

| Método | Path | Auth |
|---|---|---|
| GET | `/api/laps` | user |
| GET | `/api/laps/:id` | user |

No hay create/update — las genera el motor de reglas internamente. `data`: `[]LapResponse` o `LapResponse`:
```json
{
  "id": "uuid", "vehicle_id": "uuid", "route_id": "uuid", "lap_number": 3,
  "started_at": "...", "ended_at": "...", "duration_seconds": 4200, "allowed_duration_seconds": 3600,
  "lap_status": "LATE",
  "charge": { "id": "uuid", "amount": 5, "charge_status": "PENDING", "paid_at": null },
  "created_at": "..."
}
```
`ended_at`/`duration_seconds`/`charge` se omiten mientras la vuelta sigue `IN_PROGRESS`.

`lap_status`: `IN_PROGRESS | ON_TIME | LATE | TOO_FAST`. `charge_status`: `PENDING | PAID`.

### 2.5 Multas — `modules/fine` (base `/api/fines`, `/api/fine-types`)

| Método | Path | Auth |
|---|---|---|
| GET | `/api/fines` | **admin** |
| GET | `/api/fines/:id` | **admin** |
| PATCH | `/api/fines/:id/void` | admin |
| GET | `/api/fine-types` | user |

`FineResponse`:
```json
{
  "id": "uuid", "vehicle_id": "uuid",
  "fine_type": { "id": "uuid", "code": "SPEEDING", "name": "Exceso de velocidad", "default_amount": 50.0, "severity": "WARNING" },
  "lap_id": "uuid", "alarm_incident_id": null,
  "amount": 50.0, "fine_status": "PENDING",
  "latitude": -17.79, "longitude": -63.17, "occurred_at": "...", "notes": "72 km/h (límite 60 km/h)",
  "created_at": "..."
}
```
`fine_status`: `PENDING | PAID | VOIDED | APPEALED` (hoy solo se transiciona a `VOIDED` vía `/void`). Códigos de `fine_type.code` generados automáticamente por el motor: `OVERTAKING`, `LAP_TIME`, `PROLONGED_STOP`, `SPEEDING`.

`PATCH /api/fines/:id/void` body opcional `{ "notes": "..." }`.

> Nota: existe también `GET /api/fines/mine`, pero es exclusivo del rol dueño (fuera del alcance de este documento).

### 2.6 Geocercas — `modules/geofence` (base `/api/geofences`)

| Método | Path | Auth |
|---|---|---|
| POST / GET / PUT / PATCH `.../status` | `/api/geofences[/:id]` | user (cualquier autenticado) |

Body: `{ "name": "...", "type": "CIRCLE"|"POLYGON", "radius": 30.5, "points": [{"latitude":..,"longitude":..,"sequence":1}] }`. `CIRCLE` usa el primer punto como centro + `radius`; `POLYGON` usa todos los puntos ordenados. En `PUT`, si `points` viene, reemplaza toda la lista. Se usa para definir el `checkpoint_geofence_id` de una ruta.

### 2.7 Vehículos (reuso, genérico) — `/api/vehicles`

- `GET /api/vehicles/simple?available=true`: lista simplificada `{id, placa}` sin dispositivo instalado — útil para poblar el selector al asignar un micro a una ruta.
- `GET /api/vehicles/:id`: detalle de un vehículo (para mostrar placa/modelo al lado del pin en el mapa Admin).

### 2.8 Registro de microbús en un solo paso — `POST /api/vehicle-routes/register-micro` (admin)

`VehicleType` ahora tiene un campo `code` (varchar 8, nullable) para identificar el tipo de vehículo sin depender del UUID — el tipo "Microbús" se referencia con `code = "BUS"`. Debe existir un `VehicleType` con ese código (crearlo una vez vía `POST /api/vehicle-types` con `{"name":"Microbús","code":"BUS"}` si no existe todavía).

Este endpoint combina en una sola llamada: crear (o reusar) el `Model` bajo el `VehicleType` "BUS", crear el `Vehicle`, y opcionalmente asignarlo de una vez a una ruta con su pin. Reemplaza el flujo manual de 4 pasos para el caso de uso del panel Admin ("dar de alta un micro").

Body (`RegisterMicroRequest`):
```json
{
  "placa": "1234ABC",
  "description": "...", "year": 2020, "km_liter": 12.5,
  "chasis": "...", "color": "...", "photo_url": "...",
  "user_id": "uuid",

  "model_id": "uuid",

  "make_id": "uuid",
  "model_name": "Microbús Línea 18",

  "route_id": "uuid",
  "pin_number": "4"
}
```
Reglas:
- `placa` y `user_id` (dueño) siempre requeridos.
- Modelo: usar **`model_id`** si ya existe el modelo del micro, **o** `make_id` + `model_name` para que el backend lo busque (por nombre+marca) o lo cree bajo el `VehicleType` "BUS". Si no se puede resolver el `VehicleType` "BUS" (no está creado), responde `400` indicando crearlo primero.
- Asignación a ruta: si se manda `route_id`, `pin_number` es obligatorio (mismo `(route_id, pin_number)` único que en `POST /api/vehicle-routes`). Si se omite `route_id`, el micro queda creado sin asignar y se puede asignar después con el CRUD normal de `/api/vehicle-routes`.

`201`, `data` (`RegisterMicroResponse`):
```json
{
  "vehicle": { /* VehicleResponse, ver 2.9 (vehículos) */ },
  "vehicle_route": { "id":"uuid","vehicle_id":"uuid","route_id":"uuid","pin_number":"4","active":true,"assigned_at":"...","created_at":"...","status":true }
}
```
`vehicle_route` se omite si no se mandó `route_id`.

---

## 3. WebSocket — contrato completo

### 3.1 Canal privado (autenticado) — usado por el panel Admin para alertas en tiempo real

```
ws(s)://<host>/api/realtime/ws?token=<JWT>
```
- El JWT va como **query param**, no como header.
- Cierre `4002` razón `"No token provided"` / `"INVALID_TOKEN"` si falta o es inválido.
- Cierre `4001` razón `"TOKEN_EXPIRED"` si expiró (para que el cliente sepa que debe refrescar el token, no solo reconectar).
- Como el usuario admin, se recibe todo lo dirigido a "usuarios objetivo" además de lo global (broadcast ampliado a admins).
- El servidor hace ping cada ~54s; no envía nada el cliente, solo escucha.

### 3.2 Canal público (sin login) — útil para el mapa en vivo del panel Admin sin depender del filtro por usuario

```
ws(s)://<host>/api/realtime/public/ws?route_id=<uuid>
```
- `route_id` opcional: si se pasa, solo eventos de esa ruta; si se omite, se reciben eventos de **todas** las rutas (topic `route:all`).
- Sin JWT (puede usarse igual desde el panel Admin ya logueado, es simplemente otro socket).
- Recomendado: al conectar, pedir también `GET /api/routes/:id/live` para el snapshot inicial (el WS solo trae eventos nuevos desde que se conecta).

### 3.3 Formato del mensaje

Cada frame de texto es (o contiene, separado por `\n` si vienen varios en el mismo frame) uno o más JSON con esta forma:
```json
{ "event": "<NOMBRE_EVENTO>", "data": { } }
```

### 3.4 Catálogo de eventos

| Evento | Canal | A quién llega | `data` |
|---|---|---|---|
| `DEVICE_UPDATED` | privado | dueños del vehículo + admins | posición genérica (motos y micros) |
| `MICRO_POSITION` | público | topic de la ruta / `route:all` | posición de un micro en ruta activa |
| `LAP_COMPLETED` | privado | dueño del vehículo (+ admins) | `LapResponse` plano (igual que `/api/laps`) |
| `LAP_COMPLETED` | público | topic de la ruta / `route:all` | envuelto: `{vehicle_id, pin_number, lap: LapResponse}` |
| `NOTIFICATION` | privado | usuario destinatario (admin incluido si es el destinatario) | `NotificationResponse` (incluye `type`, hoy solo `FINE_GENERATED`) |

**`DEVICE_UPDATED`** (privado, cualquier vehículo):
```json
{ "event": "DEVICE_UPDATED", "data": {
  "imei": "861234567890123", "latitude": -17.79, "longitude": -63.17,
  "speed": 32, "course": 180, "device_time": "...", "server_time": "...",
  "battery": 85.5, "ignition": true, "satellites": 9, "category": "live" } }
```
`category` calculado server-side: `"live"` (moviéndose), `"idling"` (detenido con motor encendido), `"parked"`.

**`MICRO_POSITION`** (público) — para mover los pines en el mapa Admin en tiempo real:
```json
{ "event": "MICRO_POSITION", "data": {
  "vehicle_id": "uuid", "route_id": "uuid", "pin_number": "4",
  "latitude": -17.79, "longitude": -63.17, "speed_kmh": 32,
  "lap_number": 3, "lap_status": "IN_PROGRESS", "event_time": "..." } }
```

**`LAP_COMPLETED`** privado:
```json
{ "event": "LAP_COMPLETED", "data": { /* LapResponse, ver 2.4 */ } }
```
**`LAP_COMPLETED`** público:
```json
{ "event": "LAP_COMPLETED", "data": { "vehicle_id": "uuid", "pin_number": "4", "lap": { /* LapResponse */ } } }
```
⚠️ **Mismo nombre de evento, payload distinto según canal** — si el Admin escucha ambos canales a la vez, debe distinguir cuál está procesando.

**`NOTIFICATION`** privado — usarlo para mostrar el toast/alerta de multa nueva en el panel:
```json
{ "event": "NOTIFICATION", "data": {
  "id": "uuid", "type": "FINE_GENERATED", "title": "Nueva multa: Exceso de velocidad",
  "message": "Se generó una multa de Bs. 50.00 a tu micro por Exceso de velocidad.",
  "data": "{\"fine_id\":\"uuid\",\"amount\":50,\"fine_type\":\"SPEEDING\"}",
  "read": false, "created_at": "..." } }
```
`data.data` (anidado) es string JSON — parsear si se necesitan los campos internos.

**No existen** eventos WS dedicados por tipo de multa (`OVERTAKING_DETECTED`, `SPEEDING_DETECTED`, etc.) — esos hechos llegan como `NOTIFICATION` con `type="FINE_GENERATED"`, o se descubren consultando `/api/fines` o esperando el próximo `LAP_COMPLETED`.

---

## 4. Prompt para construir el Panel Web Admin

> Construye el panel de administración de micros. Rol: `admin` (JWT en `Authorization: Bearer`). Pantallas necesarias:
> - **Mapa en vivo** con filtro **"EN RUTA"**: sin filtro, listar todos los micros vía `GET /api/vehicle-routes` (+ datos de vehículo por `GET /api/vehicles/:id`); con el filtro activo, usar `GET /api/routes/:id/live` por cada ruta y suscribirse al WS público `route:all` (evento `MICRO_POSITION`) para actualizaciones en vivo, mostrando el `pin_number` en el pin del mapa.
> - **CRUD de Rutas** (`/api/routes`): geometría (GeoJSON), paradas, `checkpoint_geofence_id` (seleccionar de `/api/geofences`), `allowed_lap_duration_seconds`, `max_speed_kmh`, `max_stop_duration_seconds`.
> - **CRUD de Geocercas** (`/api/geofences`): tipo círculo/polígono, para definir el checkpoint de vuelta.
> - **Alta de un micro nuevo**: usar `POST /api/vehicle-routes/register-micro` (sección 2.8) para crear el vehículo (y su modelo si hace falta) y opcionalmente asignarlo a una ruta con su pin, todo en una sola llamada. Requiere que exista un `VehicleType` con `code:"BUS"` (crearlo una vez si no existe).
> - **CRUD de asignación micro↔ruta** (`/api/vehicle-routes`): para reasignar un micro existente a otra ruta/pin, o para asignar más adelante un micro creado sin ruta.
> - **CRUD de tarifas** (`/api/route-fares`): monto por vuelta y vigencia.
> - **Listado de vueltas** (`/api/laps`, solo lectura) por micro/ruta/fecha, con su cobro anidado.
> - **Listado de multas** (`/api/fines`, admin-only) con acción "anular" (`PATCH /api/fines/:id/void`).
> - **Alertas en tiempo real**: suscribirse al WS privado admin (`/api/realtime/ws?token=`) para recibir `NOTIFICATION` (multas nuevas) y `LAP_COMPLETED` de cualquier micro.

---

## 5. Cosas a decidir con negocio antes de cerrar el diseño de UI

(copiado de `docs/plan-modulo-micros.md`, sección 7 — no están resueltas del lado backend, pueden afectar flujos de UI del panel Admin)

1. Motos y micros ¿son el mismo negocio/tenant o hay que aislar datos entre operadores? Hoy el sistema es de un solo tenant.
2. El cobro por vuelta ¿es solo un registro de deuda/reporte o se integra con un medio de pago? Afecta si el panel necesita una pantalla de "cobrar"/pasarela o solo de reporte.
3. Las multas automáticas ¿se cobran directo o pasan por revisión humana antes de confirmarse? Afecta si "anular" (`/void`) es la única acción o hace falta un flujo de aprobación previo.

**Limitación técnica conocida** (puede generar falsos positivos visibles en el panel Admin): la detección de adelantamiento puede fallar justo en la "costura" donde la vuelta se cierra y reabre en el trazado GeoJSON — recomendable validar con GPS real de 2+ micros antes de mostrar/cobrar automáticamente esa multa en producción.
