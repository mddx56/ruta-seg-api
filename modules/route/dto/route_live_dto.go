package dto

import (
	"time"

	"github.com/google/uuid"
)

// LiveVehicleResponse describe a un micro con ruta activa y su última posición
// conocida (leída del cache Redis de posiciones). Usado por el canal público
// (pantalla TV, app pública) para pintar el mapa sin necesidad de login.
type LiveVehicleResponse struct {
	VehicleID          uuid.UUID `json:"vehicle_id"`
	PinNumber          string    `json:"pin_number"`
	Latitude           float64   `json:"latitude"`
	Longitude          float64   `json:"longitude"`
	SpeedKmh           int       `json:"speed_kmh"`
	LastUpdateAt       time.Time `json:"last_update_at"`
	SecondsSinceUpdate int       `json:"seconds_since_update"`
	LapNumber          *int      `json:"lap_number,omitempty"`
	LapStatus          *string   `json:"lap_status,omitempty"`
}

// EtaResponse es el tiempo estimado de llegada de un micro al punto consultado.
type EtaResponse struct {
	VehicleID      uuid.UUID `json:"vehicle_id"`
	PinNumber      string    `json:"pin_number"`
	DistanceMeters float64   `json:"distance_meters"`
	EtaSeconds     int       `json:"eta_seconds"`
}
