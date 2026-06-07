package constants

const (
	ENUM_ROLE_ADMIN    = "admin"
	ENUM_ROLE_USER     = "user"
	ENUM_ROLE_INSTALLER = "installer"

	// Literales (para mostrar en UI)
	ENUM_ROLE_ADMIN_LITERAL     = "Administrador"
	ENUM_ROLE_USER_LITERAL      = "Usuario"
	ENUM_ROLE_INSTALLER_LITERAL = "Instalador"

	ENUM_RUN_PRODUCTION = "production"
	ENUM_RUN_TESTING    = "testing"

	ENUM_PAGINATION_PER_PAGE = 10
	ENUM_PAGINATION_PAGE     = 1

	DB         = "db"
	JWTService = "JWTService"

	// Moto Enums
	ENUM_MARCA_HONDA = "Honda"
	ENUM_MARCA_HERO  = "Hero"
	ENUM_MARCA_KTM   = "KTM"
	ENUM_MARCA_VESPA = "Vespa"
	ENUM_MARCA_ATUL  = "Atul"

	ENUM_MOTO_ESTADO_NUEVA         = "Nueva"
	ENUM_MOTO_ESTADO_SEMI_NUEVA    = "SemiNueva"
	ENUM_MOTO_ESTADO_PARTICULAR    = "Particular"
	ENUM_MOTO_ESTADO_TRABAJO       = "Trabajo"
	ENUM_MOTO_ESTADO_MANTENIMIENTO = "Mantenimiento"
	ENUM_MOTO_ESTADO_SN            = "SN"

	// Device Enums
	ENUM_DEVICE_ESTADO_DISPONIBLE    = "Disponible"
	ENUM_DEVICE_ESTADO_MONITOREO     = "Monitoreo"
	ENUM_DEVICE_ESTADO_MANTENIMIENTO = "Mantenimiento"

	// Alarma Enums
	ENUM_ALARMA_ESTADO_ACTIVADA    = "ACTIVADA"
	ENUM_ALARMA_ESTADO_DESACTIVADA = "DESACTIVADA"
)

func RoleLiteral(role string) string {
	switch role {
	case ENUM_ROLE_ADMIN:
		return ENUM_ROLE_ADMIN_LITERAL
	case ENUM_ROLE_INSTALLER:
		return ENUM_ROLE_INSTALLER_LITERAL
	case ENUM_ROLE_USER:
		return ENUM_ROLE_USER_LITERAL
	default:
		return role
	}
}
