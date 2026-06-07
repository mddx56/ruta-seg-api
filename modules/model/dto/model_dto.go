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

type ModelCreateRequest struct {
	Name          string    `json:"name" binding:"required"`
	VehicleTypeID uuid.UUID `json:"vehicle_type_id" binding:"required"`
	MakeID        uuid.UUID `json:"make_id" binding:"required"`
}

type ModelUpdateRequest struct {
	ID            uuid.UUID `json:"id" binding:"required"`
	Name          string    `json:"name" binding:"omitempty"`
	VehicleTypeID uuid.UUID `json:"vehicle_type_id" binding:"omitempty"`
	MakeID        uuid.UUID `json:"make_id" binding:"omitempty"`
}

type ModelResponse struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	VehicleType *VehicleTypeInfo `json:"vehicle_type,omitempty"`
	Make        *MakeInfo        `json:"make,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	Status      bool             `json:"status"`
}

type VehicleTypeInfo struct {
	ID       uuid.UUID `json:"id"`
	TypeName string    `json:"type_name"`
}

type MakeInfo struct {
	ID       uuid.UUID `json:"id"`
	MakeName string    `json:"make_name"`
}
