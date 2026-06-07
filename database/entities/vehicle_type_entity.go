package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VehicleType struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	TypeName string    `gorm:"type:text;not null" json:"type_name"`

	Models []Model `gorm:"foreignKey:VehicleTypeID" json:"models,omitempty"`

	Timestamp
}

func (vt *VehicleType) BeforeCreate(tx *gorm.DB) (err error) {
	if vt.ID == uuid.Nil {
		vt.ID = uuid.New()
	}
	return
}
