package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Geofence struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"type:varchar(120);not null" json:"name"`
	
	Type        string    `gorm:"type:varchar(20);not null" json:"type"` // "CIRCLE" o "POLYGON"
	
	Radius      *float64  `gorm:"type:decimal(10,2)" json:"radius,omitempty"`
	
	Points      []GeofencePoint `gorm:"foreignKey:GeofenceID;constraint:OnDelete:CASCADE;" json:"points"`

	CreatedByID uuid.UUID `gorm:"type:uuid;not null" json:"created_by_id"`
	User        *User     `gorm:"foreignKey:CreatedByID" json:"user,omitempty"`
	Timestamp
}

func (g *Geofence) BeforeCreate(tx *gorm.DB) (err error) {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return
}

type GeofencePoint struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	GeofenceID uuid.UUID `gorm:"type:uuid;not null;index" json:"geofence_id"`
	
	Latitude   float64   `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude  float64   `gorm:"type:decimal(11,8);not null" json:"longitude"`
	
	Sequence   int       `gorm:"not null" json:"sequence"` 
}

func (p *GeofencePoint) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}
