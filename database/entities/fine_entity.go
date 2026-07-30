package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FineType - Catálogo de tipos de multa (adelantamiento, tiempo de vuelta, parada prolongada, exceso de velocidad)
type FineType struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Code          string    `gorm:"type:varchar(50);unique;not null" json:"code"` // OVERTAKING, LAP_TIME, PROLONGED_STOP, SPEEDING
	Name          string    `gorm:"type:varchar(100);not null" json:"name"`
	DefaultAmount float64   `gorm:"type:decimal(10,2);not null;default:0" json:"default_amount"`
	Severity      string    `gorm:"type:varchar(20);default:'WARNING'" json:"severity"`

	Timestamp
}

func (FineType) TableName() string {
	return "micros.fine_types"
}

func (ft *FineType) BeforeCreate(tx *gorm.DB) (err error) {
	if ft.ID == uuid.Nil {
		ft.ID = uuid.New()
	}
	return
}

// Fine - Multa aplicada a un micro, generada manual o automáticamente por el motor de reglas
type Fine struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	VehicleID uuid.UUID `gorm:"type:uuid;not null;index" json:"vehicle_id"`
	Vehicle   *Vehicle  `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`

	FineTypeID uuid.UUID `gorm:"type:uuid;not null" json:"fine_type_id"`
	FineType   FineType  `gorm:"foreignKey:FineTypeID" json:"fine_type"`

	LapID *uuid.UUID `gorm:"type:uuid" json:"lap_id,omitempty"`
	Lap   *Lap       `gorm:"foreignKey:LapID" json:"lap,omitempty"`

	AlarmIncidentID *uuid.UUID     `gorm:"type:uuid" json:"alarm_incident_id,omitempty"`
	AlarmIncident   *AlarmIncident `gorm:"foreignKey:AlarmIncidentID" json:"alarm_incident,omitempty"`

	Amount float64 `gorm:"type:decimal(10,2);not null" json:"amount"`

	// FineStatus: PENDING, PAID, VOIDED, APPEALED
	FineStatus string `gorm:"type:varchar(20);default:'PENDING'" json:"fine_status"`

	Latitude   float64   `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude  float64   `gorm:"type:decimal(11,8);not null" json:"longitude"`
	OccurredAt time.Time `gorm:"not null" json:"occurred_at"`

	Notes *string `gorm:"type:text" json:"notes,omitempty"`

	Timestamp
}

func (Fine) TableName() string {
	return "micros.fines"
}

func (f *Fine) BeforeCreate(tx *gorm.DB) (err error) {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return
}
