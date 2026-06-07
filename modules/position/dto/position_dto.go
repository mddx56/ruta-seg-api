package dto

import (
	"errors"
	"time"
)

const (
	// Failed
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "fallo al obtener datos del body"
	MESSAGE_FAILED_CREATE_POSITION    = "fallo al crear position"
	MESSAGE_FAILED_GET_LIST_POSITION  = "fallo al obtener lista de position"
	MESSAGE_FAILED_GET_POSITION       = "fallo al obtener position"
	MESSAGE_FAILED_DELETE_POSITION    = "fallo al eliminar position"

	// Success
	MESSAGE_SUCCESS_CREATE_POSITION   = "éxito al crear position"
	MESSAGE_SUCCESS_GET_LIST_POSITION = "éxito al obtener lista de position"
	MESSAGE_SUCCESS_GET_POSITION      = "éxito al obtener position"
	MESSAGE_SUCCESS_DELETE_POSITION   = "éxito al eliminar position"
)

var (
	ErrCreatePosition = errors.New("fallo al crear position")
	ErrGetPosition    = errors.New("fallo al obtener position")
	ErrDeletePosition = errors.New("fallo al eliminar position")
)

type PositionCreateRequest struct {
	Imei       string    `json:"device_id" binding:"required"`
	DeviceTime time.Time `json:"device_time" binding:"required"`
	Latitude   float64   `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude  float64   `json:"longitude" binding:"required,min=-180,max=180"`
	Speed      int       `json:"speed"`
	Course     int       `json:"course"`
	Attributes *string   `json:"attributes,omitempty"`
}

type PositionResponse struct {
	ID         uint64    `json:"id"`
	Imei       string    `json:"device_id"`
	ServerTime time.Time `json:"server_time"`
	DeviceTime time.Time `json:"device_time"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Speed      int       `json:"speed"`
	Course     int       `json:"course"`
	Attributes *string   `json:"attributes,omitempty"`
}

// PositionCoordinateResponse con atributos aplanados en el mismo nivel
type PositionCoordinateResponse struct {
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	ServerTime time.Time `json:"server_time"`
	DeviceTime time.Time `json:"device_time"`
	Speed      int       `json:"speed"`
	Course     int       `json:"course"`
	Battery    *float64  `json:"battery,omitempty"`
	Ignition   *bool     `json:"ignition,omitempty"`
	Satellites *int      `json:"satellites,omitempty"`
}

type DeliveryVehicleInfo struct {
	Placa       string    `json:"placa"`
	Brand       string    `json:"brand"`
	Model       string    `json:"model"`
	Type        string    `json:"type"`
	OwnerName   string    `json:"owner_name"`
	InstalledAt time.Time `json:"installed_at"`
}

type PositionsWithVehicleInfoResponse struct {
	Positions []PositionCoordinateResponse `json:"positions"`
	Vehicle   DeliveryVehicleInfo          `json:"vehicle"`
}
