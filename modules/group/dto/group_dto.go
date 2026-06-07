package dto

import (
	"time"

	"github.com/google/uuid"
)

type (
	GroupCreateRequest struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		UserID      string `json:"user_id" binding:"required"`
	}

	GroupUpdateRequest struct {
		ID          string `json:"id" binding:"required"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	GroupResponse struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description *string   `json:"description,omitempty"`
		UserID      uuid.UUID `json:"user_id"`
		IsActive    bool      `json:"is_active"`
		IsDeleted   bool      `json:"is_deleted"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	GroupAssignDeviceRequest struct {
		GroupID    uuid.UUID `json:"group_id" binding:"required"`
		DeviceIMEI string    `json:"device_imei" binding:"required"`
	}

	GroupRemoveDeviceRequest struct {
		GroupID    uuid.UUID `json:"group_id" binding:"required"`
		DeviceIMEI string    `json:"device_imei" binding:"required"`
	}
)

const (
	MESSAGE_SUCCESS_CREATE_GROUP = "Grupo creado exitosamente"
	MESSAGE_FAILED_CREATE_GROUP  = "No se pudo crear el grupo"
	MESSAGE_SUCCESS_UPDATE_GROUP = "Grupo actualizado exitosamente"
	MESSAGE_FAILED_UPDATE_GROUP  = "No se pudo actualizar el grupo"
	MESSAGE_SUCCESS_DELETE_GROUP = "Grupo eliminado exitosamente"
	MESSAGE_FAILED_DELETE_GROUP  = "No se pudo eliminar el grupo"
	MESSAGE_SUCCESS_GET_GROUPS   = "Grupos obtenidos exitosamente"
	MESSAGE_FAILED_GET_GROUPS    = "No se pudo obtener los grupos"
)
