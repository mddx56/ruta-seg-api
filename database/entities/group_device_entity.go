package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GroupDevice represents the many-to-many relationship between Groups and Devices
type GroupDevice struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	GroupID uuid.UUID `gorm:"type:uuid;not null;index" json:"group_id"`
	Group   *Group    `gorm:"foreignKey:GroupID" json:"group,omitempty"`

	DeviceIMEI string  `gorm:"type:varchar(20);not null;index" json:"device_imei"`
	Device     *Device `gorm:"foreignKey:DeviceIMEI;references:IMEI" json:"device,omitempty"`

	// Audit fields
	AssignedBy uuid.UUID `gorm:"type:uuid;not null" json:"assigned_by"`
	User       *User     `gorm:"foreignKey:AssignedBy" json:"user_assigned_by,omitempty"`

	Timestamp
}

func (GroupDevice) TableName() string {
	return "group_devices"
}

func (gd *GroupDevice) BeforeCreate(tx *gorm.DB) (err error) {
	if gd.ID == uuid.Nil {
		gd.ID = uuid.New()
	}
	return
}
