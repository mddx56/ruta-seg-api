package entities

import (
	"time"
)

type AppVersion struct {
	AppId                int       `gorm:"primaryKey;autoIncrement" json:"app_id"`
	VersionName          string    `gorm:"type:varchar(50)" json:"version_name"`
	VersionCode          string    `gorm:"type:varchar(10)" json:"version_code"`
	UrlPlaystore         string    `gorm:"type:varchar(500)" json:"url_playstore"`
	UrlApplestore        string    `gorm:"type:varchar(500)" json:"url_applestore"`
	FechaRelease         time.Time `gorm:"type:date" json:"fecha_release"`
	MiniSupportedVersion string    `gorm:"type:varchar(50)" json:"mini_supported_version"`
	IsForceUpdate        bool      `gorm:"default:false" json:"is_force_update"`
	Plataform            string    `gorm:"type:varchar(50)" json:"plataform"`

	Timestamp
}

func (AppVersion) TableName() string {
	return "app_versions"
}
