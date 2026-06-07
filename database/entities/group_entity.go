package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Group struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`

	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Relación many-to-many corregida tras eliminar GroupID de Device
	Devices []Device `gorm:"many2many:group_devices;foreignKey:ID;joinForeignKey:GroupID;References:IMEI;joinReferences:DeviceIMEI" json:"devices,omitempty"`

	Timestamp
}

func (Group) TableName() string {
	return "groups"
}

func (g *Group) BeforeCreate(tx *gorm.DB) (err error) {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return
}
