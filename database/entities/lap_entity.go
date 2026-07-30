package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Lap - Una vuelta de un micro sobre su Route, delimitada por el cruce del checkpoint de la ruta
type Lap struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	VehicleID uuid.UUID `gorm:"type:uuid;not null;index" json:"vehicle_id"`
	Vehicle   *Vehicle  `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`

	RouteID uuid.UUID `gorm:"type:uuid;not null;index" json:"route_id"`
	Route   *Route    `gorm:"foreignKey:RouteID" json:"route,omitempty"`

	LapNumber int `gorm:"not null" json:"lap_number"`

	StartedAt time.Time  `gorm:"not null" json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`

	DurationSeconds        *int `json:"duration_seconds,omitempty"`
	AllowedDurationSeconds *int `json:"allowed_duration_seconds,omitempty"`

	// LapStatus: IN_PROGRESS, ON_TIME, LATE, TOO_FAST
	LapStatus string `gorm:"type:varchar(20);default:'IN_PROGRESS'" json:"lap_status"`

	Charge *LapCharge `gorm:"foreignKey:LapID" json:"charge,omitempty"`

	// --- Estado de seguimiento del motor de infracciones (Fase 3), interno, no se expone en la API ---

	// LastMovementAt se actualiza mientras el micro se mueve; sirve para medir cuánto tiempo lleva detenido
	LastMovementAt *time.Time `json:"-"`
	// LastStopFineAt marca que ya se multó la parada actual (se limpia cuando el micro vuelve a moverse)
	LastStopFineAt *time.Time `json:"-"`
	// LastSpeedFineAt evita multar el mismo exceso de velocidad sostenido en cada posición
	LastSpeedFineAt *time.Time `json:"-"`

	// LastPolylineIndex/LastProgressMeters ubican al micro sobre el trazado de la ruta (map-matching),
	// usados para detectar adelantamientos entre micros de la misma ruta y sentido
	LastPolylineIndex  *int       `json:"-"`
	LastProgressMeters *float64   `json:"-"`
	LastProgressAt     *time.Time `json:"-"`

	// OvertakingFined evita generar más de una multa de adelantamiento por vuelta
	OvertakingFined bool `gorm:"default:false" json:"-"`

	Timestamp
}

func (Lap) TableName() string {
	return "micros.laps"
}

func (l *Lap) BeforeCreate(tx *gorm.DB) (err error) {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return
}
