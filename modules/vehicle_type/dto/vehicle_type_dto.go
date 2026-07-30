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
	TypeName string  `json:"name" binding:"required"`
	Code     *string `json:"code" binding:"omitempty,max=8"`
}

type VehicleTypeUpdateRequest struct {
	ID       uuid.UUID `json:"id" binding:"required"`
	TypeName string    `json:"name" binding:"omitempty"`
	Code     *string   `json:"code" binding:"omitempty,max=8"`
}

type VehicleTypeResponse struct {
	ID       uuid.UUID `json:"id"`
	TypeName string    `json:"name"`
	Code     *string   `json:"code,omitempty"`
}
