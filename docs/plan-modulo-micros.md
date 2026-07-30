# Plan de Implementación — Módulo "Micros" (Transporte Público)

> Basado en los mensajes del Ing. Juan Pablo Gutierrez (25/6/2026 y 1/7/2026) y en el análisis del código actual de `ruta-seg-api` (módulo de motos / "RS"). Piloto propuesto: **Línea 18** (da vueltas al primer anillo).

> **Alcance de este repositorio**: `ruta-seg-api` es solo backend. Todas las fases de este plan (incluidas las que dicen "App móvil" o "Pantalla TV") se implementan aquí únicamente como **servicios/endpoints/eventos** que esos clientes van a consumir (REST, WebSocket). El código de la app móvil, el panel web y el cliente de la pantalla TV se construye en otros repositorios, no en este.

## 1. Resumen del requerimiento

El negocio pide un nuevo módulo para transporte público ("micros"), consumido por tres clientes externos (construidos en otros repositorios, **no en `ruta-seg-api`**):

1. **Panel web de administración**: mapa con todos los micros (número en el pin), conteo de vueltas, cobro por vuelta, detección de infracciones (adelantamiento entre micros, parada prolongada, tiempo de vuelta fuera de lo permitido, exceso de velocidad).
2. **Pantalla de TV** en la oficina de los micros, mostrando la ruta marcada y los micros activos.
3. **App móvil** con dos roles:
   - **Dueño**: igual que en el módulo de motos (RS) hoy, más visualización de sus multas.
   - **Público**: sin login, ve ubicación y tiempo estimado de llegada al punto donde está consultando.

Este repositorio solo construye lo que esos tres clientes necesitan del lado del servidor: endpoints REST, eventos WebSocket y el motor de reglas que calcula vueltas/multas/cobros. No se escribe código de UI, app móvil ni cliente de TV aquí.

Regla de negocio clave aclarada el 1/7: la vista principal de Admin debe tener un filtro **"EN RUTA"** que muestre solo los micros que están trabajando en ese momento (no todos entran a ruta a la misma hora). Al activarse ese filtro es cuando se empiezan a calcular vueltas, multas, adelantamientos, etc. Aparte, el Dueño debe poder ver sus micros en todo momento (como hoy con las motos), estén o no en ruta.

## 2. Qué ya existe y qué es nuevo (hallazgos del código actual)

El proyecto sigue un patrón de módulos (`modules/<nombre>/{controller,service,repository,dto,routes.go}`) sobre Gin + GORM + `samber/do`, documentado por las skills `crear-modulo` / `crear-endpoint` y ejemplificado en `modules/make`.

**Reutilizable tal cual:**
- `Vehicle` + `VehicleType` + `Model`/`Make`: un micro es un `Vehicle` más, solo hace falta un `VehicleType` nuevo ("Microbús"). Evita duplicar todo el árbol vehículo/dispositivo.
- `Device` + `DeviceInstallation`: el hardware GPS y su historial de instalación en el vehículo ya son genéricos, no están atados a "moto".
- Ingesta de posición (HTTP `POST /api/positions` + gRPC) y **cache Redis de última posición** (`device:pos:{IMEI}`), documentado en `docs/redis-fase2-impacto.md`.
- **WebSocket** (`GET /api/realtime/ws`, `Hub`/`BroadcastPosition`) para push en tiempo real al dueño y a los admins — se puede extender con nuevos tipos de evento (vuelta completada, multa generada, etc.).
- `Geofence` / `GeofencePoint` (polígono o círculo con puntos ordenados): buena base para modelar el trazado de la ruta y/o el punto de control de vuelta, aunque **hoy no tiene módulo CRUD propio**, solo existe como campo de `AlarmRule`.
- `AlarmType` / `AlarmRule` / `AlarmIncident`: motor de reglas (límite de velocidad, geocerca, ventana horaria/día) + registro de incidentes. Es la base natural para exceso de velocidad y puede sostener el resto de infracciones si se les da un `AlarmType` propio.
- Auth (JWT access 4h + refresh token rotativo 7 días, cacheado en Redis) y `app_version` (force-update) — no requieren cambios para soportar micros, ya son agnósticos al tipo de vehículo.

**No existe y hay que construir:**
- Modelo de **Ruta** (trazado, paradas, punto de control de vuelta) — no hay nada parecido a una polilínea de ruta con paradas hoy.
- **Conteo de vueltas** y su motor de evaluación (no hay ningún worker/evaluador corriendo hoy sobre las posiciones entrantes más allá del guardado + broadcast; confirmar si `AlarmRule` realmente se evalúa en algún lado o solo existe el modelo de datos — **ver duda #1**).
- **Multas** (`Multa`/infracción con monto, estado, evidencia) — no existe tabla ni módulo. Lo más cercano es `AlarmIncident`, que registra el evento pero no tiene monto/cobro/estado de pago.
- **Cobro por vuelta** — no existe ningún concepto de tarifa/cobro en el sistema.
- **Notificaciones push** — no existe FCM, no hay tabla de tokens de dispositivo, no hay tabla de notificaciones. Hoy solo hay envío de correo transaccional (OTP, reset password).
- **Rol "Público"** sin autenticación con acceso de solo lectura a ubicación/ETA — hoy el acceso público solo existe como endpoints GET sin middleware `Authenticate` (ej. `/api/makes`), no como concepto de rol.
- **Multi-tenencia real**: no existe (el commit "tenant nuevo" fue solo un cambio de configuración de Bruno, no un cambio de modelo de datos). Si micros y motos son negocios/clientes distintos que deben estar aislados entre sí, es infraestructura nueva — **ver duda #2**.

**Riesgo detectado de paso (no es de este módulo, pero conviene resolverlo antes de escalar el sistema):** `POST /api/auth/register` está **sin autenticación** y crea usuarios con rol `admin` directamente. Recomiendo cerrarlo o protegerlo antes de sumar más superficie (más dueños, más datos sensibles de ubicación/multas).

## 3. Modelo de datos propuesto (nuevas tablas)

**Schema separado**: todas las tablas nuevas de este módulo viven en un schema Postgres propio, `micros`, en vez de `public` (donde están hoy `users`, `vehicles`, `devices`, etc.). Esto las mantiene claramente separadas de las tablas del módulo de motos. Hoy la conexión no fija ningún `search_path` ni schema (`config/database.go`), así que no hay nada que migrar: solo hace falta (a) que la migración inicial de este módulo cree el schema (`CREATE SCHEMA IF NOT EXISTS micros`), (b) que cada entidad nueva declare `TableName()` devolviendo `"micros.<tabla>"` (mismo mecanismo que ya usan `device_entity.go`, `group_entity.go`, etc., solo que hoy devuelven nombre plano sin prefijo), y (c) que el rol de base de datos tenga permisos sobre ambos schemas, ya que varias FKs cruzan de `micros` hacia `public` (`users`, `vehicles`, `devices`).

**Nombres en inglés**: siguiendo la convención ya existente en el proyecto (`vehicles`, `positions`, `alarm_incidents`...), todas las tablas y columnas nuevas van en inglés.

Siguiendo el patrón existente (UUID PK, `Timestamp` embebido, `status bool`):

| Entidad (tabla) | Campos clave | Estado | Notas |
|---|---|---|---|
| `Route` (`micros.routes`) | name, description, map_color, checkpoint_geofence_id, allowed_lap_duration_seconds, max_speed_kmh, max_stop_duration_seconds, geometry (jsonb), active, created_by | ✅ implementado + CRUD (`modules/route`) | "Línea 18 - Primer Anillo" ya está sembrada. `geometry` guarda el GeoJSON completo del trazado (fijo, cargado una sola vez vía seeder desde `database/seeders/data/linea-18.geojson`) — **no** hay tabla de puntos (`RoutePoint` se descartó): el backend lo parsea (`pkg/geo`) para map-matching. `checkpoint_geofence_id` referencia una `Geofence` (círculo, en `public`) que marca inicio/fin de vuelta. `max_speed_kmh`/`max_stop_duration_seconds` (Fase 3) son opcionales y configurables por ruta desde el panel Admin. |
| `RouteStop` (`micros.route_stops`) | route_id, name, latitude, longitude, sequence | ✅ implementado (parte de `modules/route`) | Paradas de referencia, usadas para calcular ETA hacia un punto consultado por el público. |
| `VehicleRoute` (`micros.vehicle_routes`) | vehicle_id, route_id, pin_number, active, assigned_at | ✅ implementado + CRUD (`modules/vehicle_route`) | Asigna un micro a una ruta y define el número que se muestra en el pin del mapa (no necesariamente la placa). |
| `Lap` (`micros.laps`) | vehicle_id, route_id, lap_number, started_at, ended_at, duration_seconds, allowed_duration_seconds, lap_status (in_progress/on_time/late/too_fast), charge (relación 1:1) | ✅ implementado + motor de reglas (`modules/lap`) | Se genera/cierra automáticamente cuando la posición de un micro entra al checkpoint de su ruta. Solo lectura vía API (`GET /api/laps`); lo escribe el motor. |
| `FineType` (`micros.fine_types`) | code (OVERTAKING, LAP_TIME, PROLONGED_STOP, SPEEDING), name, default_amount, severity | ✅ sembrado + lectura (`modules/fine`) | Catálogo, igual de espíritu a `AlarmType`. Las 4 filas ya están cargadas vía `FineTypeSeeder`. |
| `Fine` (`micros.fines`) | vehicle_id, fine_type_id, lap_id (nullable), alarm_incident_id (nullable), amount, fine_status (pending/paid/voided/appealed), latitude, longitude, occurred_at, notes | ✅ los 4 tipos generan automáticamente (`modules/fine` + `modules/lap`) | `LAP_TIME` (Fase 2) y `OVERTAKING`/`PROLONGED_STOP`/`SPEEDING` (Fase 3) ya se generan solos desde el motor de reglas, todos vía `FineService.GenerateFine`. `alarm_incident_id` queda sin usar por ahora (ver nota sobre `AlarmRule` en la sección 6, Fase 3). |
| `RouteFare` (`micros.route_fares`) | route_id, amount_per_lap, effective_from | ✅ implementado + CRUD (`modules/route_fare`) | Tarifa vigente para el cobro automático por vuelta. |
| `LapCharge` (`micros.lap_charges`) | lap_id, amount, charge_status (pending/paid), paid_at | ✅ generación automática, expuesto anidado en `Lap.charge` | Un cobro generado por cada `Lap` completada, usando la `RouteFare` vigente al momento del cierre. |
| `UserDeviceToken` (`micros.user_device_tokens`) | user_id, token, platform | ✅ implementado + CRUD (`modules/notification`) | Registro de tokens FCM del dispositivo móvil, vía `POST/DELETE /api/notifications/device-tokens`. |
| `Notification` (`micros.notifications`) | user_id, type, title, message, read, data (json), created_at | ✅ implementado + generación automática | Historial de notificaciones (`GET /api/notifications`), generado por `NotificationService.Notify` — hoy usado por `modules/fine` para `FINE_GENERATED`, entregado por WebSocket + intento de push (FCM). |

`VehicleType` (en `public`) gana una columna `code` (varchar(8), nullable) para referenciar tipos sin depender del UUID, y gana una fila nueva ("Microbús", `code:"BUS"`) — **la columna ya existe (migrada), la fila pendiente de crear** vía `POST /api/vehicle-types {"name":"Microbús","code":"BUS"}` o automáticamente al usar `POST /api/vehicle-routes/register-micro` (falla con mensaje claro si el tipo "BUS" todavía no existe). Ese endpoint nuevo registra un microbús en un solo paso: crea/reusa el `Model` bajo el tipo "BUS", crea el `Vehicle` y opcionalmente lo asigna a una ruta con su pin. No se toca el rol de usuario: "dueño" sigue siendo el rol `user` existente, la propiedad se resuelve igual que hoy vía `Vehicle.UserID` — un mismo dueño podría tener motos y micros sin cambios de modelo.

También ya se construyó el módulo `Geofence` (CRUD en `modules/geofence`), que no existía antes de este trabajo y era prerrequisito de `Route.checkpoint_geofence_id`.

## 4. Requisitos funcionales

Cada subsección describe qué necesita un cliente, pero lo que se construye en `ruta-seg-api` es siempre el **servicio del lado del servidor** (endpoint REST o evento WebSocket) que ese cliente consume — nunca la pantalla en sí.

### 4.1 Servicios para el Panel Admin (Web)
- **RF-01** Endpoint/eventos para listar todos los micros con posición + número de pin en tiempo real (reusa WebSocket + cache Redis de posición ya existente; el pin sale de `VehicleRoute.pin_number` ✅ ya expuesto en `GET /api/vehicle-routes`).
- **RF-02** Endpoint que exponga cuáles micros están "EN RUTA" (posición reciente y coincidente con el trazado de su ruta asignada). Al quedar "en ruta", ese micro pasa a tener vueltas/multas/adelantamientos calculados.
- **RF-03** Endpoint de conteo de vueltas por micro (vuelta actual, vueltas del día, hora de inicio/fin de cada vuelta) — sobre la tabla `Lap`.
- **RF-04** Endpoint de cobros generados por vuelta, con estado (pendiente/pagado) y totales por micro/día — sobre `LapCharge`.
- **RF-05** Endpoint de multas generadas (tipo, monto, hora, ubicación) con acción de anular/aprobar (revisión humana antes de cobrar, a definir — ver duda #5) — sobre `Fine`.
- **RF-06** CRUD para definir/editar una ruta (geometría, paradas, checkpoint de vuelta, duración permitida de vuelta y tarifa por vuelta) — **✅ ya implementado** (`modules/route`, `modules/geofence`; falta CRUD de `RouteFare`). El panel Admin es quien dibuja/importa la geometría; acá solo se recibe y guarda.
- **RF-07** Eventos en tiempo real (WebSocket) de: adelantamiento detectado, parada prolongada, exceso de velocidad, vuelta fuera de tiempo.

### 4.2 Servicios para la Pantalla de TV (oficina de los micros)
- **✅ RF-08** Canal público de WebSocket `GET /api/realtime/public/ws?route_id=<uuid>` (sin login) — ver arquitectura de tiempo real en la sección 6, Fase 6. Complementado por `GET /api/routes/{id}/live` para el estado inicial al cargar la pantalla.

### 4.3 Servicios para la App móvil — rol Dueño
- **RF-09** Mismo backend de registro/login y gestión de vehículos que hoy usa el módulo de motos (RS), reusado tal cual para micros (un micro es un `Vehicle` con `VehicleType` "Microbús").
- **RF-10** Endpoint/eventos de posición de sus propios micros en tiempo real, estén o no "en ruta" (no depende del filtro de Admin) — ya cubierto por `GET /api/vehicles/my` + WebSocket existente.
- **✅ RF-11** `GET /api/fines/mine` (autenticado): historial de multas de los vehículos del usuario. Se aprovechó para cerrar un hueco de seguridad: `GET /api/fines` (todas) y `GET /api/fines/{id}` ahora son **solo admin** — antes cualquier usuario autenticado podía ver las multas de cualquier otro dueño.
- **✅ RF-12** Al generarse una multa (`FineService.GenerateFine`), se notifica al dueño del micro (+ admins) por el WebSocket privado existente (evento `FINE_GENERATED`), igual patrón que `BroadcastPosition`. Push real vía FCM sigue pendiente de Fase 4.

### 4.4 Servicios para la App móvil — rol Público
- **✅ RF-13** `GET /api/routes/{id}/live` (público): lista de micros con posición reciente (≤2 min) en esa ruta, con pin, posición, velocidad y vuelta en curso. Complementado por el canal WebSocket público de la Fase 6 para actualizaciones en tiempo real sin polling.
- **✅ RF-14** `GET /api/routes/{id}/eta?lat=&lon=`(público): ETA de cada micro en vivo de la ruta al punto consultado, usando map-matching (`pkg/geo`) sobre `Route.geometry` — calcula el avance de cada micro y del punto consultado sobre la misma polilínea/sentido, con envoltura de vuelta (loop) si el micro ya pasó ese punto en la vuelta actual. Velocidad usada: la actual del micro si es significativa, si no, un promedio derivado de `allowed_lap_duration_seconds` (o 20 km/h por defecto).

### 4.5 Motor de reglas (backend, corre sobre el flujo de posiciones)
- **RF-15** Conteo de vueltas: al detectar que la posición del micro entra al `checkpoint_geofence` de su ruta (con una distancia/tiempo mínimo desde la última entrada para evitar duplicados), cerrar el `Lap` anterior y abrir uno nuevo.
- **RF-16** Cálculo de duración de vuelta vs. `allowed_lap_duration_seconds`; si excede, generar `Fine` tipo `LAP_TIME`.
- **RF-17** Cobro automático: al cerrar un `Lap`, generar `LapCharge` según `RouteFare` vigente.
- **✅ RF-18** Detección de adelantamiento: por cada posición, se proyecta el micro sobre la polilínea de su ruta/sentido (`pkg/geo.BestProjection`, map-matching) para obtener su avance acumulado (metros desde el inicio del trazado). Se compara contra el avance de los demás micros con vuelta en curso en la misma ruta y mismo sentido: si un micro que arrancó su vuelta **después** ya tiene más avance que uno que arrancó **antes** (con un margen de tolerancia de 30 m para el ruido de GPS), se genera `Fine` tipo `OVERTAKING` contra el que adelantó. Como máximo una multa de este tipo por vuelta por micro.
- **✅ RF-19** Detección de parada prolongada: se rastrea `Lap.last_movement_at` (se actualiza mientras la velocidad reportada supera 5 km/h); si pasa más tiempo que `Route.max_stop_duration_seconds` (default 300 s si la ruta no lo configura) sin movimiento, se genera `Fine` tipo `PROLONGED_STOP` (una sola vez por parada; se resetea cuando el micro vuelve a moverse).
- **✅ RF-20** Exceso de velocidad: **no** se reusó `AlarmRule`/`AlarmIncident` (ver nota en sección 6, Fase 3 — ese subsistema nunca se llegó a construir, ni siquiera su CRUD). En su lugar, cada ruta tiene su propio `max_speed_kmh` opcional; si la velocidad reportada lo supera, se genera `Fine` tipo `SPEEDING` (con cooldown de 5 min para no repetir la multa mientras el exceso se mantiene sostenido).
- **✅ RF-21** Notificar en tiempo real cada evento: `LAP_COMPLETED` y `FINE_GENERATED` van al dueño del micro + admins por el WebSocket privado existente; `MICRO_POSITION` y `LAP_COMPLETED` (versión pública, sin datos sensibles) van al canal público por ruta vía Redis Pub/Sub (ver Fase 6). Solo falta el **push** real vía FCM (Fase 4) — hoy la "notificación" es únicamente el evento WebSocket.

## 5. Requisitos no funcionales
- Los cálculos de vueltas/adelantamiento/parada deben tolerar el ruido normal del GPS (jitter, pérdida de señal) sin generar falsos positivos — necesita un umbral de distancia/tiempo, no comparación exacta.
- El filtro "EN RUTA" y las alertas deben sentirse en tiempo real (mismo SLA que el tracking de motos hoy, vía WebSocket), no por polling.
- El trazado de ruta y el catálogo de multas deben ser editables sin tocar código (vía panel Admin), para poder mapear nuevas líneas después del piloto de la Línea 18.
- Cualquier multa generada automáticamente debe guardar evidencia (posición, hora, velocidad) para poder ser auditada o apelada.

## 6. Fases propuestas

Todas las fases son trabajo de backend (endpoints, eventos, motor de reglas). Ningún ítem implica escribir código de panel web, app móvil o cliente de TV — esos se construyen en sus propios repositorios y solo consumen lo que se lista acá.

1. **✅ Fase 0 — Piloto Línea 18 (mapeo)**: geometría real de la Línea 18 recibida como GeoJSON y cargada vía seeder (`database/seeders/seeds/route_seed.go`) en `Route.geometry`. *(Se descartó levantar un trazado propio: la geometría ya venía dada y es fija, ver sección 3.)*
2. **✅ Fase 1 — Modelo de datos + CRUD base**: entidades de la sección 3 (schema `micros`), CRUD de `Geofence` (no existía) y `Route`/`RouteStop`, y CRUD de `VehicleRoute` para asignar micros a una ruta con su número de pin. Falta solo: seed de `VehicleType` "Microbús".
3. **✅ Fase 2 — Motor de vueltas y cobro**: RF-15 a RF-17 implementados en `modules/lap` (motor), `modules/fine` (catálogo + multas) y `modules/route_fare` (tarifas). El motor (`LapService.EvaluatePosition`) se engancha de forma no bloqueante a la ingesta de posiciones existente (`modules/position`, tanto HTTP como gRPC): detecta el cruce del checkpoint vía `pkg/geo` (Haversine para `CIRCLE`, ray-casting para `POLYGON`), cierra/abre `Lap`, clasifica `on_time/late/too_fast`, genera `Fine` tipo `LAP_TIME` si corresponde y `LapCharge` según la `RouteFare` vigente. Probado en vivo de punta a punta (dos posiciones GPS reales → vuelta cerrada, multa y cobro generados correctamente).
4. **✅ Fase 3 — Motor de infracciones**: RF-18 a RF-20 implementados dentro del mismo `LapService.EvaluatePosition` (`evaluateProlongedStop`, `evaluateSpeeding`, `evaluateOvertaking`), reusando `FineService.GenerateFine`. Decisión de diseño importante: se investigó reusar `AlarmRule`/`AlarmIncident` como proponía el plan original, pero se confirmó que ese subsistema **nunca se terminó de construir** — `AlarmIncidentService`/`Repository`/`Controller` son interfaces vacías sin ningún método, y no hay ningún worker que evalúe `AlarmRule` contra posiciones entrantes (duda #1 resuelta: no existe evaluador). Reconstruir ese subsistema genérico completo (matching por dispositivo/geocerca/ventana horaria/día de la semana) era un esfuerzo separado y mayor al alcance de "micros", así que se optó por un límite de velocidad simple y directo por `Route` en su lugar. Para adelantamiento se agregó map-matching (`pkg/geo`: `ParsePolylines`, `ProjectOntoPolyline`, `BestProjection`) sobre la geometría ya cargada de la ruta. Probado con tests unitarios exhaustivos (parada prolongada, velocidad con cooldown, adelantamiento con vehículo simulado) y en vivo para velocidad/parada con el vehículo de prueba real.
   - **Limitación conocida, a validar con datos reales de la Línea 18**: el adelantamiento compara avance solo entre micros proyectados sobre la misma polilínea (mismo "sentido" del GeoJSON); si el trazado real tiene ambigüedad en la costura donde termina/empieza la vuelta (el punto de progreso "envuelve" de vuelta a 0 en el checkpoint), podría haber falsos positivos/negativos justo en ese tramo. Recomendado probar con GPS real de 2+ micros antes de activar el cobro/multa automática de adelantamiento en producción.
5. **✅ Fase 4 — Notificaciones**: módulo `modules/notification` centraliza "avisarle algo a un usuario" en un solo punto (`NotificationService.Notify`): persiste en `Notification` (historial en `GET /api/notifications`), lo manda por el WebSocket privado existente, y si el usuario tiene un token registrado (`POST /api/notifications/device-tokens`, RF de gestión de tokens), dispara el push. `modules/fine` ya usa este único punto para RF-12 en vez de llamar directo al WebSocket. Cliente FCM propio y liviano en `pkg/fcm` (API HTTP v1 de Firebase + OAuth2 de cuenta de servicio, sin traer el SDK completo `firebase-admin-go`). **Configuración pendiente del lado de operaciones**: hace falta un proyecto de Firebase real y su JSON de cuenta de servicio (`FCM_CREDENTIALS_FILE` o `FCM_CREDENTIALS_JSON` en `.env`) para que el push llegue de verdad a un celular — sin eso, el sistema sigue funcionando igual (persistencia + WebSocket), solo no llega la notificación si la app está cerrada. Probado en vivo de punta a punta (sin credenciales FCM reales): multa generada → notificación persistida correctamente para el dueño real del vehículo, confirmado consultando la base de datos directamente.
6. **✅ Fase 5 — Endpoints para la App móvil**: `GET /api/fines/mine` (RF-11, con el fix de seguridad de que `GET /api/fines` pasó a ser admin-only) y notificación en tiempo real `FINE_GENERATED` al dueño (RF-12, reusando el WebSocket privado existente).
7. **✅ Fase 6 — Endpoints para la Pantalla de TV + canal público en tiempo real**: se construyó una arquitectura nueva y más robusta que un simple polling, tal como se pidió:
   - **Canal público de WebSocket** `GET /api/realtime/public/ws?route_id=<uuid>` (sin login) — extiende el `Hub` existente (`providers/websocket/hub.go`) con un sistema de **topics** (`route:<routeID>` o `route:all`), separado del ruteo por usuario/admin que ya existía.
   - **Redis Pub/Sub para funcionar con varias instancias del servidor**: el motor de reglas (`modules/lap/service`) publica cada posición de un micro "en ruta" y cada `LAP_COMPLETED` al canal Redis `micros:route-events` (`providers/websocket/route_events.go`, `RouteEventPublisher`). Cada instancia del servidor corre un suscriptor (`StartRouteEventSubscriber`, arrancado junto al hub en `providers/core.go`) que reenvía esos eventos a sus clientes locales conectados a ese topic. Así, sin importar a qué instancia esté conectada la pantalla TV o la app pública, recibe los eventos igual — a diferencia del `Hub` original, que solo llegaba a clientes de la misma instancia.
   - **`GET /api/routes/{id}/live`** (REST, público): snapshot inicial de los micros en vivo de una ruta (posición desde el cache Redis de posiciones ya existente + vuelta en curso), para cuando un cliente recién se conecta antes de que lleguen eventos por WebSocket.
   - Probado en vivo de punta a punta: cliente WebSocket público conectado, posición GPS real enviada por HTTP → evento `MICRO_POSITION` recibido por el cliente en menos de 3 segundos, pasando por Redis.

## 7. Dudas / preguntas abiertas para el Ing. Juan Pablo

1. ~~¿Existe hoy algún proceso que evalúe `AlarmRule` contra las posiciones entrantes, o el modelo de alarmas está definido pero sin motor de ejecución corriendo?~~ **Resuelto**: no existe. Ni el evaluador ni el CRUD de `AlarmIncident` (están vacíos). Se implementó el límite de velocidad de forma independiente por `Route` en vez de reusar ese subsistema — si más adelante quieren el motor de alarmas genérico para motos también, es trabajo aparte, no cubierto por este plan.
2. Motos y micros: ¿son el mismo negocio/tenant o hay que aislar datos entre distintos clientes/operadores de micros? Hoy el sistema es de un solo tenant por instalación.
3. El cobro por vuelta: ¿es solo un registro de deuda/reporte, o debe integrarse con algún medio de pago (efectivo en oficina, billetera, pasarela de pago)?
4. Las multas generadas automáticamente (adelantamiento, tiempo de vuelta, etc.) ¿se cobran directo o pasan por una revisión humana antes de confirmarse?
5. Rol "Público": ¿totalmente sin login, o se permite opcionalmente crear cuenta para guardar líneas favoritas/recibir notificaciones de su micro?
6. ¿El "adelantamiento" se define solo por invertir el orden entre dos micros de la misma vuelta, o hay una distancia mínima de por medio (p.ej. no cuenta si van a menos de X metros)? **Implementado con una tolerancia de 30 m** (constante `overtakingProximityToleranceMeters` en `modules/lap/service`) para no generar falsos positivos por ruido de GPS entre micros muy cercanos — ajustable si en la práctica resulta muy laxo/estricto.
7. ~~Para el piloto de la Línea 18, ¿ya existe el trazado del primer anillo en algún formato (KML/GPX, Google My Maps), o se levanta recorriendo la ruta con un dispositivo?~~ **Resuelto**: se entregó `linea-18.geojson` (2 `MultiLineString`, uno por sentido) y ya está cargado en `Route.geometry`.
8. Pantalla TV: ¿corre en un navegador normal (Chrome en un Smart TV / mini-PC) o necesita ser una app dedicada?
9. Push notifications: ¿ya existe un proyecto de Firebase para la app (o para RS/motos) del que se pueda sacar el JSON de cuenta de servicio, o hay que crear uno nuevo? Sin eso, `FCM_CREDENTIALS_FILE`/`FCM_CREDENTIALS_JSON` quedan sin configurar y las notificaciones solo llegan por WebSocket (con la app abierta), no como push al celular.
