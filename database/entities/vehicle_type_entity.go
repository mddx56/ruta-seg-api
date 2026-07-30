package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VehicleType struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	TypeName string    `gorm:"type:text;not null" json:"type_name"`
	// Code identifica el tipo de vehículo de forma estable (ej. "BUS" para microbuses),
	// para poder referenciarlo desde otros módulos sin depender del UUID.
	Code *string `gorm:"type:varchar(8)" json:"code,omitempty"`

	Models []Model `gorm:"foreignKey:VehicleTypeID" json:"models,omitempty"`

	Timestamp
}

func (vt *VehicleType) BeforeCreate(tx *gorm.DB) (err error) {
	if vt.ID == uuid.Nil {
		vt.ID = uuid.New()
	}
	return
}
