package entities

import (
	"gorm.io/gorm"
)

type Device struct {
	IMEI string `gorm:"primaryKey;type:varchar(20);not null" json:"imei"`

	Timestamp

	Model string `gorm:"type:varchar(50);default:'GT06'" json:"model"`

	Protocol        *string `gorm:"type:varchar(50)" json:"protocol,omitempty"`
	SimPhoneNumber  *string `gorm:"type:varchar(20)" json:"sim_phone_number,omitempty"`
	SimICCID        *string `gorm:"column:sim_icc_id;type:varchar(30)" json:"sim_icc_id,omitempty"`
	SimProvider     *string `gorm:"type:varchar(7)" json:"sim_provider,omitempty"`
	APNConf         *string `gorm:"type:jsonb" json:"apn_conf,omitempty"`
	FirmwareVersion *string `gorm:"type:varchar(50)" json:"firmware_version,omitempty"`
	RemoteIP        *string `gorm:"type:inet" json:"remote_ip,omitempty"`

	// Auditoría interna
	UserCreator *string `gorm:"type:uuid" json:"-"`
	Creator     *User   `gorm:"foreignKey:UserCreator" json:"-"`
	UserUpdater *string `gorm:"type:uuid" json:"-"`
	Updater     *User   `gorm:"foreignKey:UserUpdater" json:"-"`
	Batch       *string `gorm:"type:uuid" json:"-"`

	Positions     []Position           `gorm:"foreignKey:Imei" json:"-"`
	Installations []DeviceInstallation `gorm:"foreignKey:Imei;references:IMEI" json:"installations,omitempty"`
	GroupDevices  []GroupDevice        `gorm:"foreignKey:DeviceIMEI;references:IMEI" json:"group_devices,omitempty"`
	Vehicles      []Vehicle            `gorm:"-" json:"vehicles,omitempty"`
}

func (Device) TableName() string {
	return "devices"
}

func (d *Device) BeforeCreate(tx *gorm.DB) (err error) {
	return
}
