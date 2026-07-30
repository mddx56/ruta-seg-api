package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserDeviceToken - Token FCM del dispositivo móvil de un usuario, usado para push notifications
type UserDeviceToken struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`

	Token    string `gorm:"type:text;not null;uniqueIndex" json:"token"`
	Platform string `gorm:"type:varchar(20);not null" json:"platform"` // android, ios

	Timestamp
}

func (UserDeviceToken) TableName() string {
	return "micros.user_device_tokens"
}

func (t *UserDeviceToken) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

// Notification - Notificación entregada a un usuario (multa nueva, adelantamiento, etc.), vía WebSocket y/o push
type Notification struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User   *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`

	Type    string  `gorm:"type:varchar(50);not null" json:"type"`
	Title   string  `gorm:"type:varchar(150);not null" json:"title"`
	Message string  `gorm:"type:text;not null" json:"message"`
	Data    *string `gorm:"type:jsonb" json:"data,omitempty"`
	Read    bool    `gorm:"default:false" json:"read"`

	Timestamp
}

func (Notification) TableName() string {
	return "micros.notifications"
}

func (n *Notification) BeforeCreate(tx *gorm.DB) (err error) {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return
}
