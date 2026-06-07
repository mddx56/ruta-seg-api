package dto

import (
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

type VehicleTypeCreateRequest struct {
	TypeName string `json:"name" binding:"required"`
}

type VehicleTypeUpdateRequest struct {
	ID       uuid.UUID `json:"id" binding:"required"`
	TypeName string    `json:"name" binding:"omitempty"`
}

type VehicleTypeResponse struct {
	ID       uuid.UUID `json:"id"`
	TypeName string    `json:"name"`
}
