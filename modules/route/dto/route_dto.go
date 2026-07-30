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

type RouteStopRequest struct {
	Name      string  `json:"name" binding:"required"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Sequence  int     `json:"sequence"`
}

type RouteCreateRequest struct {
	Name                      string             `json:"name" binding:"required"`
	Description               *string            `json:"description,omitempty"`
	MapColor                  *string            `json:"map_color,omitempty"`
	CheckpointGeofenceID      *uuid.UUID         `json:"checkpoint_geofence_id,omitempty"`
	AllowedLapDurationSeconds *int               `json:"allowed_lap_duration_seconds,omitempty"`
	MaxSpeedKmh               *int               `json:"max_speed_kmh,omitempty"`
	MaxStopDurationSeconds    *int               `json:"max_stop_duration_seconds,omitempty"`
	Geometry                  *string            `json:"geometry,omitempty"`
	Stops                     []RouteStopRequest `json:"stops,omitempty" binding:"omitempty,dive"`
}

type RouteUpdateRequest struct {
	ID                        uuid.UUID          `json:"id" binding:"-"`
	Name                      string             `json:"name" binding:"omitempty"`
	Description               *string            `json:"description,omitempty"`
	MapColor                  *string            `json:"map_color,omitempty"`
	CheckpointGeofenceID      *uuid.UUID         `json:"checkpoint_geofence_id,omitempty"`
	AllowedLapDurationSeconds *int               `json:"allowed_lap_duration_seconds,omitempty"`
	MaxSpeedKmh               *int               `json:"max_speed_kmh,omitempty"`
	MaxStopDurationSeconds    *int               `json:"max_stop_duration_seconds,omitempty"`
	Geometry                  *string            `json:"geometry,omitempty"`
	Active                    *bool              `json:"active,omitempty"`
	Stops                     []RouteStopRequest `json:"stops,omitempty" binding:"omitempty,dive"`
}

type RouteStopResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Sequence  int       `json:"sequence"`
}

type RouteResponse struct {
	ID                        uuid.UUID           `json:"id"`
	Name                      string              `json:"name"`
	Description               *string             `json:"description,omitempty"`
	MapColor                  *string             `json:"map_color,omitempty"`
	CheckpointGeofenceID      *uuid.UUID          `json:"checkpoint_geofence_id,omitempty"`
	AllowedLapDurationSeconds *int                `json:"allowed_lap_duration_seconds,omitempty"`
	MaxSpeedKmh               *int                `json:"max_speed_kmh,omitempty"`
	MaxStopDurationSeconds    *int                `json:"max_stop_duration_seconds,omitempty"`
	Geometry                  *string             `json:"geometry,omitempty"`
	Active                    bool                `json:"active"`
	Stops                     []RouteStopResponse `json:"stops"`
	CreatedByID               uuid.UUID           `json:"created_by_id"`
	CreatedAt                 time.Time           `json:"created_at"`
	Status                    bool                `json:"status"`
}
