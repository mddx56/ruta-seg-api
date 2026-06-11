Crea un módulo CRUD completo en este proyecto Go (Gin + GORM + samber/do).

## Paso 1 — Recopilar información

Antes de escribir cualquier archivo, pregunta al usuario:

1. **Nombre del módulo** (snake_case, singular). Ejemplo: `vehicle_type`, `alarm_rule`, `skill`
2. **Campos de la entidad** (además de `id`, `status`, `created_at`, `updated_at`). Ejemplo: `name string requerido`, `description string opcional`
3. **¿Requiere autenticación en GET?** (sí/no — por defecto sí)
4. **¿Solo admin puede crear/editar/eliminar?** (sí/no — por defecto sí)

Espera las respuestas antes de continuar.

## Paso 2 — Leer archivos de referencia

Lee estos archivos para entender los patrones exactos del proyecto antes de escribir código:

- `modules/make/routes.go`
- `modules/make/repository/make_repository.go`
- `modules/make/service/make_service.go`
- `modules/make/controller/make_controller.go`
- `modules/make/dto/make_dto.go`
- `database/entities/make_entity.go`
- `database/entities/common.go`
- `database/migrations/20260216202000_create_app_versions_table.go`
- `providers/core.go`

## Paso 3 — Crear archivos

Usa los datos recopilados. Las variables son:
- `{name}` = nombre en snake_case (ej: `vehicle_type`)
- `{Name}` = PascalCase (ej: `VehicleType`)
- `{TIMESTAMP}` = fecha/hora actual en formato `YYYYMMDDHHMMSS`

### 1. `database/entities/{name}_entity.go`

```go
package entities

import (
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type {Name} struct {
    ID   uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    // campos del usuario aquí
    Timestamp
}

func (e *{Name}) BeforeCreate(_ *gorm.DB) (err error) {
    if e.ID == uuid.Nil {
        e.ID = uuid.New()
    }
    return
}
```

### 2. `database/migrations/{TIMESTAMP}_create_{name}s_table.go`

```go
package migrations

import (
    "github.com/Caknoooo/go-gin-clean-starter/database"
    "github.com/Caknoooo/go-gin-clean-starter/database/entities"
    "gorm.io/gorm"
)

func init() {
    database.RegisterMigration("{TIMESTAMP}_create_{name}s_table", Up{TIMESTAMP}Create{Name}sTable, Down{TIMESTAMP}Create{Name}sTable)
}

func Up{TIMESTAMP}Create{Name}sTable(db *gorm.DB) error {
    return db.AutoMigrate(&entities.{Name}{})
}

func Down{TIMESTAMP}Create{Name}sTable(db *gorm.DB) error {
    return db.Migrator().DropTable(&entities.{Name}{})
}
```

### 3. `modules/{name}/dto/{name}_dto.go`

Incluye constantes de mensajes en español y los structs:
- `{Name}CreateRequest` — campos requeridos con `binding:"required"`
- `{Name}UpdateRequest` — campo `ID uuid.UUID` + campos opcionales con `binding:"omitempty"`
- `{Name}Response` — ID, campos del dominio, Status

### 4. `modules/{name}/repository/{name}_repository.go`

Interface con: `Create`, `Update`, `ChangeStatus`, `FindAll`, `FindByID`.

Constructor recibe `*do.Injector`, obtiene `*gorm.DB` con `do.MustInvokeNamed[*gorm.DB](injector, constants.DB)`.

`ChangeStatus` usa:
```go
r.db.WithContext(ctx).Model(&entities.{Name}{}).Where("id = ?", id).Update("status", gorm.Expr("NOT status")).Error
```

`FindAll` filtra `status = true` y ordena por el campo principal.

### 5. `modules/{name}/service/{name}_service.go`

Interface con los mismos 5 métodos trabajando con DTOs.

Constructor recibe `*do.Injector`, obtiene repo con `do.MustInvoke[repository.{Name}Repository](injector)`.

Helper privado `toResponse(e entities.{Name}) dto.{Name}Response`.

### 6. `modules/{name}/controller/{name}_controller.go`

Interface + struct con 5 handlers Gin: `Create`, `Update`, `ChangeStatus`, `FindAll`, `FindByID`.

Constructor recibe `*do.Injector`, obtiene service con `do.MustInvoke[service.{Name}Service](injector)`.

Cada handler con comentarios Swagger (`@Summary`, `@Tags`, `@Router`, `@Security BearerAuth`).

Patrón de error en Create/Update:
- Bind error → `http.StatusBadRequest`
- UUID parse error → `http.StatusBadRequest`
- Service error → `http.StatusInternalServerError`
- Éxito Create → `http.StatusCreated`
- Éxito resto → `http.StatusOK`

### 7. `modules/{name}/routes.go`

```go
package {name}

import (
    "github.com/Caknoooo/go-gin-clean-starter/middlewares"
    authService "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
    "github.com/Caknoooo/go-gin-clean-starter/modules/{name}/controller"
    "github.com/Caknoooo/go-gin-clean-starter/modules/{name}/repository"
    "github.com/Caknoooo/go-gin-clean-starter/modules/{name}/service"
    "github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
    "github.com/gin-gonic/gin"
    "github.com/samber/do"
)

func RegisterRoutes(router *gin.Engine, injector *do.Injector) {
    jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

    do.Provide(injector, repository.New{Name}Repository)
    do.Provide(injector, service.New{Name}Service)
    do.Provide(injector, controller.New{Name}Controller)

    ctrl := do.MustInvoke[controller.{Name}Controller](injector)

    g := router.Group("/api/{name}s")
    {
        g.POST("",           middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), ctrl.Create)
        g.GET("",            middlewares.Authenticate(jwtService), ctrl.FindAll)
        g.GET("/:id",        middlewares.Authenticate(jwtService), ctrl.FindByID)
        g.PUT("/:id",        middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), ctrl.Update)
        g.PATCH("/:id/status", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), ctrl.ChangeStatus)
    }
}
```

Ajusta los middlewares según las respuestas del Paso 1.

## Paso 4 — Registrar en cmd/main.go

Agrega el import y la llamada a `RegisterRoutes` junto a los otros módulos:

```go
import (
    // ...
    {name}module "github.com/Caknoooo/go-gin-clean-starter/modules/{name}"
)

// en main():
{name}module.RegisterRoutes(server, injector)
```

## Paso 5 — Verificar

Ejecuta `go build ./...`. Si hay errores, corrígelos antes de reportar éxito.

Reporta al usuario:
- Lista de archivos creados/modificados
- Endpoints generados con método HTTP, ruta y permisos
