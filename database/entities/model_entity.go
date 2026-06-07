package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	ModelName string    `gorm:"type:text;not null" json:"model_name"`

	VehicleTypeID uuid.UUID    `gorm:"type:uuid;not null" json:"vehicle_type_id"`
	VehicleType   *VehicleType `gorm:"foreignKey:VehicleTypeID" json:"vehicle_type,omitempty"`

	MakeID uuid.UUID `gorm:"type:uuid;not null" json:"make_id"`
	Make   *Make     `gorm:"foreignKey:MakeID" json:"make,omitempty"`

	Vehicles []Vehicle `gorm:"foreignKey:ModelID" json:"vehicles,omitempty"`

	Timestamp
}

func (m *Model) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}
