package dto

import "time"

// VehiclePositionResponse — último punto conocido de un vehículo
// Agrupa por vehículo (no por IMEI), usando el dispositivo actualmente instalado.
type VehiclePositionResponse struct {
	VehicleID  string    `json:"vehicle_id"`
	Placa      string    `json:"placa"`
	Imei       string    `json:"imei"` // Dispositivo activo actualmente
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Speed      int       `json:"speed"`
	Course     int       `json:"course"`
	DeviceTime time.Time `json:"device_time"`
	ServerTime time.Time `json:"server_time"`
	Attributes *string   `json:"attributes,omitempty"`
}

// VehicleHistoryResponse — historial de un vehículo (no de un dispositivo)
// Idéntico a HistoryResponse pero con vehicle_id en lugar de imei.
type VehicleHistoryResponse struct {
	VehicleID     string          `json:"vehicle_id"`
	Placa         string          `json:"placa,omitempty"`
	Date          string          `json:"date"`
	TotalDistance float64         `json:"total_distance_km"`
	TotalDuration string          `json:"total_duration"`
	MaxSpeed      int             `json:"max_speed"`
	Events        []TimelineEvent `json:"events"`
}
