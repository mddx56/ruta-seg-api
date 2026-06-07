package dto

const (
	MESSAGE_SUCCESS_GET_STATS = "Estadísticas obtenidas correctamente"
	MESSAGE_FAILED_GET_STATS  = "Fallo al obtener estadísticas"
)

type DashboardStatsResponse struct {
	TotalUsers                 int64 `json:"total_users"`
	UsersWithInstallations     int64 `json:"users_with_installations"`
	SuspendedUsers             int64 `json:"suspended_users"`
	TotalDevices               int64 `json:"total_devices"`
	DevicesWithInstallation    int64 `json:"devices_with_installation"`
	DevicesWithoutInstallation int64 `json:"devices_without_installation"`
	TotalVehicles              int64 `json:"total_vehicles"`
	TotalGroups                int64 `json:"total_groups"`
}
