Crea un módulo CRUD completo en este proyecto Go siguiendo exactamente el patrón del módulo `modules/make/`.

## Argumento

El nombre del módulo a crear se pasa como argumento al invocar el skill. Ejemplo: `/crear-modulo vehiculo_tipo`

Si no se pasa argumento, pregunta al usuario el nombre antes de continuar.

## Contexto del proyecto

- Framework: Gin + GORM + samber/do (DI)
- Módulo de referencia: `modules/make/` — léelo completo antes de empezar
- Entidades base: `database/entities/common.go` — el `Timestamp` embebido da `created_at`, `updated_at`, `status`
- Patrón de migración: `database/migrations/20260216202000_create_app_versions_table.go`
- Registro de rutas: `cmd/main.go` — agregar `{módulo}.RegisterRoutes(server, injector)`

## Archivos que debes crear

Para un módulo llamado `$NAME` (snake_case), crea exactamente estos archivos:

1. **`database/entities/{name}_entity.go`**
   - Struct con UUID primary key, campos del dominio, embed `Timestamp`
   - Hook `BeforeCreate` para asignar UUID si es nil

2. **`database/migrations/{TIMESTAMP}_create_{name}s_table.go`**
   - Timestamp = fecha/hora actual en formato `YYYYMMDDHHMMSS`
   - Usa `db.AutoMigrate(&entities.{Name}{})` en Up
   - Usa `db.Migrator().DropTable(...)` en Down

3. **`modules/{name}/dto/{name}_dto.go`**
   - Constantes: `MESSAGE_SUCCESS`, `MESSAGE_CREATED`, `MESSAGE_UPDATED`, `MESSAGE_FAILED_BAD_REQUEST`, `MESSAGE_FAILED_INVALID_ID`, `MESSAGE_INTERNAL_SERVER_ERROR`
   - Structs: `{Name}CreateRequest`, `{Name}UpdateRequest` (con campo `ID uuid.UUID`), `{Name}Response`

4. **`modules/{name}/repository/{name}_repository.go`**
   - Interface con: `Create`, `Update`, `ChangeStatus`, `FindAll`, `FindByID`
   - Constructor: `func New{Name}Repository(injector *do.Injector) ({Name}Repository, error)`
   - `ChangeStatus`: usa `Model(&entities.{Name}{}).Where("id = ?", id).Update("status", gorm.Expr("NOT status"))`
   - `FindAll`: filtra `status = true`, ordena por nombre

5. **`modules/{name}/service/{name}_service.go`**
   - Interface con los mismos 5 métodos que el repo pero trabajando con DTOs
   - Constructor: `func New{Name}Service(injector *do.Injector) ({Name}Service, error)`
   - Helper privado `toResponse(e entities.{Name}) dto.{Name}Response`

6. **`modules/{name}/controller/{name}_controller.go`**
   - Interface + struct con los 5 handlers de Gin
   - Constructor: `func New{Name}Controller(injector *do.Injector) ({Name}Controller, error)`
   - Cada handler con comentario Swagger `// @Summary`, `// @Tags`, `// @Router`

7. **`modules/{name}/routes.go`**
   - Package `{name}`
   - Registra repo → service → controller en el injector
   - Rutas bajo `/api/{name}s`:
     - `POST /` — solo admin
     - `GET /` — autenticado
     - `GET /:id` — autenticado
     - `PUT /:id` — solo admin
     - `PATCH /:id/status` — solo admin

## Archivos que debes modificar

8. **`cmd/main.go`**
   - Agrega import del módulo
   - Agrega `{name}module.RegisterRoutes(server, injector)` junto a los otros módulos

## Campos del dominio

Antes de crear los archivos, pregunta al usuario:
> ¿Qué campos (además de ID, status, created_at, updated_at) debe tener la entidad `{Name}`?

Espera la respuesta y úsalos en la entidad, DTOs y respuestas.

## Verificación final

Ejecuta `go build ./...` al terminar. Si hay errores de compilación, corrígelos antes de reportar éxito.
