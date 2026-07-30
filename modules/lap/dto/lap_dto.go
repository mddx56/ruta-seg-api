package dto

import (
	"time"

	"github.com/google/uuid"
)

const (
	MESSAGE_SUCCESS               = "Éxito"
	MESSAGE_FAILED_INVALID_ID     = "ID inválido"
	MESSAGE_INTERNAL_SERVER_ERROR = "Error interno del servidor"
)

type LapChargeResponse struct {
	ID           uuid.UUID  `json:"id"`
	Amount       float64    `json:"amount"`
	ChargeStatus string     `json:"charge_status"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
}

// MicroLiveEvent es el payload publicado al canal público de micros en ruta
// (WebSocket público / TV) por cada posición de un micro con ruta activa.
type MicroLiveEvent struct {
	VehicleID uuid.UUID `json:"vehicle_id"`
	RouteID   uuid.UUID `json:"route_id"`
	PinNumber string    `json:"pin_number"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	SpeedKmh  int       `json:"speed_kmh"`
	LapNumber *int      `json:"lap_number,omitempty"`
	LapStatus *string   `json:"lap_status,omitempty"`
	EventTime time.Time `json:"event_time"`
}

// MicroLapCompletedEvent es el payload publicado al cerrar una vuelta, para que la
// pantalla TV / app pública puedan mostrar el conteo sin necesidad de autenticarse.
type MicroLapCompletedEvent struct {
	VehicleID uuid.UUID   `json:"vehicle_id"`
	PinNumber string      `json:"pin_number"`
	Lap       LapResponse `json:"lap"`
}

type LapResponse struct {
	ID                     uuid.UUID          `json:"id"`
	VehicleID              uuid.UUID          `json:"vehicle_id"`
	RouteID                uuid.UUID          `json:"route_id"`
	LapNumber              int                `json:"lap_number"`
	StartedAt              time.Time          `json:"started_at"`
	EndedAt                *time.Time         `json:"ended_at,omitempty"`
	DurationSeconds        *int               `json:"duration_seconds,omitempty"`
	AllowedDurationSeconds *int               `json:"allowed_duration_seconds,omitempty"`
	LapStatus              string             `json:"lap_status"`
	Charge                 *LapChargeResponse `json:"charge,omitempty"`
	CreatedAt              time.Time          `json:"created_at"`
}
