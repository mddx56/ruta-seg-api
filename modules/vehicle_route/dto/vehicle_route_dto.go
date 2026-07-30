package dto

import (
	"time"

	vehicledto "github.com/Caknoooo/go-gin-clean-starter/modules/vehicle/dto"
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

type VehicleRouteCreateRequest struct {
	VehicleID  uuid.UUID  `json:"vehicle_id" binding:"required"`
	RouteID    uuid.UUID  `json:"route_id" binding:"required"`
	PinNumber  string     `json:"pin_number" binding:"required"`
	AssignedAt *time.Time `json:"assigned_at,omitempty"`
}

type VehicleRouteUpdateRequest struct {
	ID        uuid.UUID  `json:"id" binding:"-"`
	RouteID   *uuid.UUID `json:"route_id,omitempty"`
	PinNumber string     `json:"pin_number" binding:"omitempty"`
	Active    *bool      `json:"active,omitempty"`
}

type VehicleRouteResponse struct {
	ID         uuid.UUID `json:"id"`
	VehicleID  uuid.UUID `json:"vehicle_id"`
	RouteID    uuid.UUID `json:"route_id"`
	PinNumber  string    `json:"pin_number"`
	Active     bool      `json:"active"`
	AssignedAt time.Time `json:"assigned_at"`
	CreatedAt  time.Time `json:"created_at"`
	Status     bool      `json:"status"`
}

// RegisterMicroRequest registra un microbús en un solo paso: crea (o reusa) el
// Model bajo el VehicleType code=BUS, crea el Vehicle, y opcionalmente lo asigna
// a una ruta con su número de pin.
type RegisterMicroRequest struct {
	Placa       string    `json:"placa" binding:"required"`
	Description *string   `json:"description,omitempty"`
	Year        *int      `json:"year,omitempty"`
	KmLiter     *float64  `json:"km_liter,omitempty"`
	Chassis     *string   `json:"chasis,omitempty"`
	Color       *string   `json:"color,omitempty"`
	PhotoURL    *string   `json:"photo_url,omitempty"`
	UserID      uuid.UUID `json:"user_id" binding:"required"`

	// Usar un Model ya existente...
	ModelID *uuid.UUID `json:"model_id,omitempty"`
	// ...o crear uno nuevo (o reusar uno con el mismo nombre+marca) bajo el VehicleType "BUS"
	MakeID    *uuid.UUID `json:"make_id,omitempty"`
	ModelName *string    `json:"model_name,omitempty"`

	// Asignación a ruta, opcional (se puede hacer después vía /api/vehicle-routes)
	RouteID   *uuid.UUID `json:"route_id,omitempty"`
	PinNumber *string    `json:"pin_number,omitempty"`
}

type RegisterMicroResponse struct {
	Vehicle      vehicledto.VehicleResponse `json:"vehicle"`
	VehicleRoute *VehicleRouteResponse      `json:"vehicle_route,omitempty"`
}
