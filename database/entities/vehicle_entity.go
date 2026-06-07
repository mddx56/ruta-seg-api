package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Vehicle struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Placa       string    `gorm:"type:text;unique;not null" json:"placa"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	Year        *int      `gorm:"type:int" json:"year,omitempty"`
	KmLiter     *float64  `gorm:"type:float" json:"km_liter,omitempty"`
	Chassis     *string   `gorm:"type:text;unique" json:"chasis,omitempty"`
	Color       *string   `gorm:"type:text" json:"color,omitempty"`
	PhotoURL    *string   `gorm:"type:text" json:"photo_url,omitempty"`

	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`

	ModelID uuid.UUID `gorm:"type:uuid;not null" json:"model_id"`
	Model   *Model    `gorm:"foreignKey:ModelID" json:"model,omitempty"`

	Devices []Device `gorm:"many2many:device_vehicles;" json:"devices,omitempty"`

	Installations []DeviceInstallation `gorm:"foreignKey:VehicleID" json:"installations,omitempty"`

	Timestamp
}

func (v *Vehicle) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return
}
