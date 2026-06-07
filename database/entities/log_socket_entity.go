package entities

import (
	"time"
)

type LogSocket struct {
	LogID uint `gorm:"primaryKey"`

	Payload string `gorm:"type:text;not null"`

	CreatedAt time.Time `gorm:"type:timestamp with time zone" json:"created_at"`
}

func (LogSocket) TableName() string {
	return "log_sockets"
}
