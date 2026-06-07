package dto

import (
	"time"

	"github.com/google/uuid"
)

type DeviceInstallationCreateRequest struct {
	Imei          string    `json:"imei" binding:"required"` // IMEI
	VehicleID     uuid.UUID `json:"vehicle_id" binding:"required"`
	InstallReason *string   `json:"install_reason,omitempty"`
}

// DeviceInstallationQuickCreateRequest se usa para la app móvil:
// permite registrar una instalación solo con IMEI y chasis del vehículo.
type DeviceInstallationQuickCreateRequest struct {
	Imei         string  `json:"imei" binding:"required"`   // IMEI del dispositivo
	Chassis      string  `json:"chasis" binding:"required"` // Chasis del vehículo
	InstallReason *string `json:"install_reason,omitempty"`
}

type DeviceInstallationResponse struct {
	InstallationID uuid.UUID  `json:"installation_id"`
	Imei           string     `json:"imei"` // IMEI
	VehicleID      uuid.UUID  `json:"vehicle_id"`
	InstalledAt    time.Time  `json:"installed_at"`
	RemovedAt      *time.Time `json:"removed_at,omitempty"`
	InstallReason  *string    `json:"install_reason,omitempty"`
	RemovalReason  *string    `json:"removal_reason,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Status         bool       `json:"status"`
}

type DeviceInstallationUninstallRequest struct {
	RemovalReason *string `json:"removal_reason,omitempty"`
}

// DeviceInstallationMineResponse es un response enriquecido para la app móvil.
type DeviceInstallationMineResponse struct {
	InstallationID uuid.UUID  `json:"installation_id"`
	Imei           string     `json:"imei"`
	InstalledAt    time.Time  `json:"installed_at"`
	RemovedAt      *time.Time `json:"removed_at,omitempty"`
	InstallReason  *string    `json:"install_reason,omitempty"`
	RemovalReason  *string    `json:"removal_reason,omitempty"`
	WorkOrderID    *string    `json:"work_order_id,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	Status         bool       `json:"status"`

	Device  *DeviceInstallationDeviceInfo  `json:"device,omitempty"`
	Vehicle *DeviceInstallationVehicleInfo `json:"vehicle,omitempty"`
}

type DeviceInstallationDeviceInfo struct {
	IMEI           string  `json:"imei"`
	Model          string  `json:"model,omitempty"`
	SimPhoneNumber *string `json:"sim_phone_number,omitempty"`
	SimICCID       *string `json:"sim_iccid,omitempty"`
	SimProvider    *string `json:"sim_provider,omitempty"`
	Protocol       *string `json:"protocol,omitempty"`
	FirmwareVersion *string `json:"firmware_version,omitempty"`
}

type DeviceInstallationVehicleInfo struct {
	ID          uuid.UUID `json:"id"`
	Placa       string    `json:"placa"`
	Chassis     *string   `json:"chasis,omitempty"`
	Description *string   `json:"description,omitempty"`
	Year        *int      `json:"year,omitempty"`
	KmLiter     *float64  `json:"km_liter,omitempty"`
	Color       *string   `json:"color,omitempty"`
	PhotoURL    *string   `json:"photo_url,omitempty"`

	Model *DeviceInstallationModelInfo `json:"model,omitempty"`
	Owner *DeviceInstallationOwnerInfo `json:"owner,omitempty"`
}

type DeviceInstallationModelInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Make *DeviceInstallationMakeInfo `json:"make,omitempty"`
}

type DeviceInstallationMakeInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type DeviceInstallationOwnerInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}
