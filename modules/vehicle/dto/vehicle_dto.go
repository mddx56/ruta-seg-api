package dto

import (
	"time"

	"github.com/google/uuid"
)

const (
	MESSAGE_SUCCESS                  = "Operación realizada correctamente"
	MESSAGE_CREATED                  = "Vehículo registrado correctamente en el sistema"
	MESSAGE_UPDATED                  = "La información del vehículo ha sido actualizada"
	MESSAGE_DELETED                  = "Vehículo eliminado correctamente"
	MESSAGE_FAILED_BAD_REQUEST       = "Los datos proporcionados no son válidos"
	MESSAGE_FAILED_INVALID_ID        = "No se encontró el vehículo con el identificador proporcionado"
	MESSAGE_INTERNAL_SERVER_ERROR    = "Ocurrió un error interno al procesar la solicitud"
	MESSAGE_FAILED_DUPLICATE_PLACA   = "Ya existe un vehículo registrado con esta placa"
	MESSAGE_FAILED_DUPLICATE_CHASSIS = "Ya existe un vehículo registrado con este número de chasis"
)

type VehicleCreateRequest struct {
	Placa       string    `json:"placa" binding:"required"`
	Description *string   `json:"description"`
	Year        *int      `json:"year"`
	KmLiter     *float64  `json:"km_liter"`
	Chassis     *string   `json:"chasis"`
	Color       *string   `json:"color"`
	PhotoURL    *string   `json:"photo_url"`
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	ModelID     uuid.UUID `json:"model_id" binding:"required"`
}

type VehicleUpdateRequest struct {
	ID          uuid.UUID `json:"id" binding:"required"`
	Placa       string    `json:"placa" binding:"omitempty"`
	Description *string   `json:"description"`
	Year        *int      `json:"year"`
	KmLiter     *float64  `json:"km_liter"`
	Chassis     *string   `json:"chasis"`
	Color       *string   `json:"color"`
	PhotoURL    *string   `json:"photo_url"`
	UserID      uuid.UUID `json:"user_id" binding:"omitempty"`
	ModelID     uuid.UUID `json:"model_id" binding:"omitempty"`
}

type VehicleResponse struct {
	ID          uuid.UUID  `json:"id"`
	Placa       string     `json:"placa"`
	Description *string    `json:"description,omitempty"`
	Year        *int       `json:"year,omitempty"`
	KmLiter     *float64   `json:"km_liter,omitempty"`
	Chassis     *string    `json:"chasis"`
	Color       *string    `json:"color,omitempty"`
	PhotoURL    *string    `json:"photo_url,omitempty"`
	User        *UserInfo  `json:"user,omitempty"`
	Model       *ModelInfo `json:"model,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Status      bool       `json:"status"`

	ActiveInstallation *VehicleInstallationInfo `json:"active_installation,omitempty"`
	Group              *GroupInfo               `json:"group,omitempty"`
}

type VehicleInstallationInfo struct {
	InstallationID uuid.UUID `json:"installation_id"`
	DeviceIMEI     string    `json:"device_imei"`
	InstalledAt    time.Time `json:"installed_at"`
}

type GroupInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type UserInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

type ModelInfo struct {
	ID        uuid.UUID `json:"id"`
	ModelName string    `json:"name"`
	Make      *MakeInfo `json:"make,omitempty"`
}

type MakeInfo struct {
	ID       uuid.UUID `json:"id"`
	MakeName string    `json:"name"`
}

type VehicleTypeInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type VehicleFullDeviceInfo struct {
	IMEI           string  `json:"imei"`
	Model          string  `json:"model,omitempty"`
	SimPhoneNumber *string `json:"sim_phone_number,omitempty"`
	SimProvider    *string `json:"sim_provider,omitempty"`
}

type VehicleFullInstallationInfo struct {
	InstallationID uuid.UUID            `json:"installation_id"`
	InstalledAt    time.Time            `json:"installed_at"`
	InstallReason  *string              `json:"install_reason,omitempty"`
	Device         *VehicleFullDeviceInfo `json:"device,omitempty"`
}

type VehicleFullResponse struct {
	Vehicle VehicleResponse `json:"vehicle"`
	VehicleType *VehicleTypeInfo `json:"vehicle_type,omitempty"`
	ActiveInstallation *VehicleFullInstallationInfo `json:"active_installation,omitempty"`
	AvailableForInstallation bool `json:"available_for_installation"`
}

type VehicleSimpleResponse struct {
	ID        uuid.UUID `json:"id"`
	Placa     string    `json:"placa"`
	Chassis   *string   `json:"chasis"`
	ModelName string    `json:"model_name"`
	MakeName  string    `json:"make_name"`
}
