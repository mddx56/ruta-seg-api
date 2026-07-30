package dto

import (
	"time"

	"github.com/google/uuid"
)

const (
	MESSAGE_SUCCESS               = "Éxito"
	MESSAGE_UPDATED               = "Actualizado exitosamente"
	MESSAGE_FAILED_BAD_REQUEST    = "Petición incorrecta"
	MESSAGE_FAILED_INVALID_ID     = "ID inválido"
	MESSAGE_INTERNAL_SERVER_ERROR = "Error interno del servidor"
)

type FineTypeResponse struct {
	ID            uuid.UUID `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	DefaultAmount float64   `json:"default_amount"`
	Severity      string    `json:"severity"`
}

type FineVoidRequest struct {
	Notes *string `json:"notes,omitempty"`
}

type FineResponse struct {
	ID              uuid.UUID        `json:"id"`
	VehicleID       uuid.UUID        `json:"vehicle_id"`
	FineType        FineTypeResponse `json:"fine_type"`
	LapID           *uuid.UUID       `json:"lap_id,omitempty"`
	AlarmIncidentID *uuid.UUID       `json:"alarm_incident_id,omitempty"`
	Amount          float64          `json:"amount"`
	FineStatus      string           `json:"fine_status"`
	Latitude        float64          `json:"latitude"`
	Longitude       float64          `json:"longitude"`
	OccurredAt      time.Time        `json:"occurred_at"`
	Notes           *string          `json:"notes,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
}
