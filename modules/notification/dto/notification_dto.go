package dto

import (
	"time"

	"github.com/google/uuid"
)

const (
	MESSAGE_SUCCESS               = "Éxito"
	MESSAGE_CREATED               = "Creado exitosamente"
	MESSAGE_UPDATED               = "Actualizado exitosamente"
	MESSAGE_FAILED_BAD_REQUEST    = "Petición incorrecta"
	MESSAGE_FAILED_INVALID_ID     = "ID inválido"
	MESSAGE_INTERNAL_SERVER_ERROR = "Error interno del servidor"
)

type RegisterDeviceTokenRequest struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform" binding:"required,oneof=android ios"`
}

type UnregisterDeviceTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type NotificationResponse struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Data      *string   `json:"data,omitempty"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}
