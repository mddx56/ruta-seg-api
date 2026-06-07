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

type MakeCreateRequest struct {
	Name string `json:"name" binding:"required"`
}

type MakeUpdateRequest struct {
	ID   uuid.UUID `json:"id" binding:"required"`
	Name string    `json:"name" binding:"omitempty"`
}

type MakeResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	// CreatedAt time.Time `json:"created_at"`
	// Status    bool      `json:"status"`
}
