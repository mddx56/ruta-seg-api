package dto

import (
	"time"

	"github.com/google/uuid"
)

const (
	MESSAGE_SUCCESS               = "Éxito"
	MESSAGE_CREATED               = "Creado exitosamente"
	MESSAGE_UPDATED               = "Actualizado exitosamente"
	MESSAGE_DELETED               = "Eliminado exitosamente"
	MESSAGE_FAILED_BAD_REQUEST    = "Petición incorrecta"
	MESSAGE_FAILED_INVALID_ID     = "ID inválido"
	MESSAGE_INTERNAL_SERVER_ERROR = "Error interno del servidor"
)

type GeofencePointRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Sequence  int     `json:"sequence"`
}

type GeofenceCreateRequest struct {
	Name   string                 `json:"name" binding:"required"`
	Type   string                 `json:"type" binding:"required,oneof=CIRCLE POLYGON"`
	Radius *float64               `json:"radius,omitempty"`
	Points []GeofencePointRequest `json:"points" binding:"required,min=1,dive"`
}

type GeofenceUpdateRequest struct {
	ID     uuid.UUID              `json:"id" binding:"-"`
	Name   string                 `json:"name" binding:"omitempty"`
	Type   string                 `json:"type" binding:"omitempty,oneof=CIRCLE POLYGON"`
	Radius *float64               `json:"radius,omitempty"`
	Points []GeofencePointRequest `json:"points,omitempty" binding:"omitempty,dive"`
}

type GeofencePointResponse struct {
	ID        uuid.UUID `json:"id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Sequence  int       `json:"sequence"`
}

type GeofenceResponse struct {
	ID          uuid.UUID               `json:"id"`
	Name        string                  `json:"name"`
	Type        string                  `json:"type"`
	Radius      *float64                `json:"radius,omitempty"`
	Points      []GeofencePointResponse `json:"points"`
	CreatedByID uuid.UUID               `json:"created_by_id"`
	CreatedAt   time.Time               `json:"created_at"`
	Status      bool                    `json:"status"`
}
