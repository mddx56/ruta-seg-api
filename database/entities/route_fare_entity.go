package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RouteFare - Tarifa vigente por vuelta completada en una Route
type RouteFare struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	RouteID uuid.UUID `gorm:"type:uuid;not null;index" json:"route_id"`
	Route   *Route    `gorm:"foreignKey:RouteID" json:"route,omitempty"`

	AmountPerLap  float64   `gorm:"type:decimal(10,2);not null" json:"amount_per_lap"`
	EffectiveFrom time.Time `gorm:"not null" json:"effective_from"`

	Timestamp
}

func (RouteFare) TableName() string {
	return "micros.route_fares"
}

func (rf *RouteFare) BeforeCreate(tx *gorm.DB) (err error) {
	if rf.ID == uuid.Nil {
		rf.ID = uuid.New()
	}
	return
}

// LapCharge - Cobro generado automáticamente al cerrar una Lap, según la RouteFare vigente
type LapCharge struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	LapID uuid.UUID `gorm:"type:uuid;not null;index" json:"lap_id"`
	Lap   *Lap      `gorm:"foreignKey:LapID" json:"lap,omitempty"`

	Amount float64 `gorm:"type:decimal(10,2);not null" json:"amount"`

	// ChargeStatus: PENDING, PAID
	ChargeStatus string     `gorm:"type:varchar(20);default:'PENDING'" json:"charge_status"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`

	Timestamp
}

func (LapCharge) TableName() string {
	return "micros.lap_charges"
}

func (lc *LapCharge) BeforeCreate(tx *gorm.DB) (err error) {
	if lc.ID == uuid.Nil {
		lc.ID = uuid.New()
	}
	return
}
