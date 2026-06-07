package entities

import (
	"time"

	"github.com/google/uuid"
)

type DeviceInstallation struct {
	InstallationID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"installation_id"`

	// Relaciones (Invariables)
	// Imei now refers to the device ID
	Imei      string    `gorm:"type:varchar(20);not null;index;column:imei" json:"imei"`
	VehicleID uuid.UUID `gorm:"type:uuid;not null" json:"vehicle_id"`

	// El Factor Tiempo (Crucial para el historial)
	InstalledAt time.Time  `gorm:"not null;default:now()" json:"installed_at"`
	RemovedAt   *time.Time `json:"removed_at,omitempty"` // NULL significa "Actualmente Instalado"

	// Datos de Contexto (Enriquecimiento)
	UserCreationID *uuid.UUID `gorm:"type:uuid" json:"user_creation_id,omitempty"` // Usuario que realizó el registro
	WorkOrderID    *string    `gorm:"type:varchar(50)" json:"work_order_id,omitempty"`

	// Motivo del cambio
	InstallReason *string `gorm:"type:varchar(50)" json:"install_reason,omitempty"` // 'new_client', 'repair', 'replacement'
	RemovalReason *string `gorm:"type:varchar(50)" json:"removal_reason,omitempty"` // 'cancelled', 'device_failure', 'sold_vehicle'

	Notes *string `gorm:"type:text" json:"notes,omitempty"` // Observaciones: "Se escondió antena bajo el tablero"

	Timestamp

	// Relations
	Device       *Device  `gorm:"foreignKey:Imei;references:IMEI" json:"device,omitempty"`
	Vehicle      *Vehicle `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	UserCreation *User    `gorm:"foreignKey:UserCreationID" json:"user_creation,omitempty"`
}

func (DeviceInstallation) TableName() string {
	return "device_installations"
}
