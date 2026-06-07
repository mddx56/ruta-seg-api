package entities

import (
	"time"
)

type Timestamp struct {
	CreatedAt time.Time `gorm:"type:timestamp with time zone" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp with time zone" json:"updated_at"`
	Status    bool      `gorm:"default:true" json:"status"`
}

type Authorization struct {
	Token string `json:"token" binding:"required"`
	Role  string `json:"role" binding:"required,oneof=user admin"`
}
