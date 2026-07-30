package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Route - Línea/recorrido de un micro (ej. "Línea 18 - Primer Anillo")
type Route struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"type:varchar(120);not null" json:"name"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	MapColor    *string   `gorm:"type:varchar(20)" json:"map_color,omitempty"`

	CheckpointGeofenceID *uuid.UUID `gorm:"type:uuid" json:"checkpoint_geofence_id,omitempty"`
	Geofence             *Geofence  `gorm:"foreignKey:CheckpointGeofenceID" json:"checkpoint_geofence,omitempty"`

	AllowedLapDurationSeconds *int `gorm:"type:integer" json:"allowed_lap_duration_seconds,omitempty"`

	// MaxSpeedKmh - Límite de velocidad de la ruta, usado por el motor de infracciones (Fase 3)
	MaxSpeedKmh *int `gorm:"type:integer" json:"max_speed_kmh,omitempty"`

	// MaxStopDurationSeconds - Tiempo máximo detenido "en ruta" antes de considerarse parada prolongada
	MaxStopDurationSeconds *int `gorm:"type:integer" json:"max_stop_duration_seconds,omitempty"`

	Active bool `gorm:"default:true" json:"active"`

	CreatedByID uuid.UUID `gorm:"type:uuid;not null" json:"created_by_id"`
	User        *User     `gorm:"foreignKey:CreatedByID" json:"user,omitempty"`

	// Geometry - Trazado de la ruta en formato GeoJSON (FeatureCollection/LineString), fijo por línea
	Geometry *string `gorm:"type:jsonb" json:"geometry,omitempty"`

	Stops []RouteStop `gorm:"foreignKey:RouteID;constraint:OnDelete:CASCADE;" json:"stops,omitempty"`

	Timestamp
}

func (Route) TableName() string {
	return "micros.routes"
}

func (r *Route) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}

// RouteStop - Parada de referencia de la ruta, usada para calcular ETA al público
type RouteStop struct {
	ID      uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	RouteID uuid.UUID `gorm:"type:uuid;not null;index" json:"route_id"`

	Name string `gorm:"type:varchar(120);not null" json:"name"`

	Latitude  float64 `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude float64 `gorm:"type:decimal(11,8);not null" json:"longitude"`

	Sequence int `gorm:"not null" json:"sequence"`
}

func (RouteStop) TableName() string {
	return "micros.route_stops"
}

func (s *RouteStop) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
