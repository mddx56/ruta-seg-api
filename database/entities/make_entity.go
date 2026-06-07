package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Make struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	MakeName string    `gorm:"type:text;unique;not null" json:"make_name"`

	Models []Model `gorm:"foreignKey:MakeID" json:"models,omitempty"`

	Timestamp
}

func (m *Make) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}
