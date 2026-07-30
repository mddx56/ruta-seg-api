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

type RouteFareCreateRequest struct {
	RouteID       uuid.UUID  `json:"route_id" binding:"required"`
	AmountPerLap  float64    `json:"amount_per_lap" binding:"required"`
	EffectiveFrom *time.Time `json:"effective_from,omitempty"`
}

type RouteFareUpdateRequest struct {
	ID            uuid.UUID  `json:"id" binding:"-"`
	AmountPerLap  *float64   `json:"amount_per_lap,omitempty"`
	EffectiveFrom *time.Time `json:"effective_from,omitempty"`
}

type RouteFareResponse struct {
	ID            uuid.UUID `json:"id"`
	RouteID       uuid.UUID `json:"route_id"`
	AmountPerLap  float64   `json:"amount_per_lap"`
	EffectiveFrom time.Time `json:"effective_from"`
	CreatedAt     time.Time `json:"created_at"`
	Status        bool      `json:"status"`
}
