package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VehicleRoute - Asignación de un micro (Vehicle) a una Route, con el número que se muestra en el pin del mapa
type VehicleRoute struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	VehicleID uuid.UUID `gorm:"type:uuid;not null;index" json:"vehicle_id"`
	Vehicle   *Vehicle  `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`

	RouteID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_route_pin_number" json:"route_id"`
	Route   *Route    `gorm:"foreignKey:RouteID" json:"route,omitempty"`

	PinNumber string `gorm:"type:varchar(10);not null;uniqueIndex:idx_route_pin_number" json:"pin_number"`

	Active     bool      `gorm:"default:true" json:"active"`
	AssignedAt time.Time `gorm:"not null" json:"assigned_at"`

	Timestamp
}

func (VehicleRoute) TableName() string {
	return "micros.vehicle_routes"
}

func (vr *VehicleRoute) BeforeCreate(tx *gorm.DB) (err error) {
	if vr.ID == uuid.Nil {
		vr.ID = uuid.New()
	}
	return
}
