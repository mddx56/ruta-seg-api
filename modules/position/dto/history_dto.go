package dto

import "time"

// HistoryResponse es lo que recibe el Frontend
type HistoryResponse struct {
	IMEI          string          `json:"imei"`
	Date          string          `json:"date"`
	TotalDistance float64         `json:"total_distance_km"` // Resumen del día
	TotalDuration string          `json:"total_duration"`
	MaxSpeed      int             `json:"max_speed"`
	Events        []TimelineEvent `json:"events"` // Lista ordenada: Parada -> Viaje -> Parada
}

// TimelineEvent representa un bloque en la línea de tiempo
type TimelineEvent struct {
	Type      string      `json:"type"` // "trip" o "stop" o "gap"
	StartTime time.Time   `json:"start_time"`
	EndTime   time.Time   `json:"end_time"`
	Duration  string      `json:"duration"` // "1h 30m"
	Data      interface{} `json:"data"`     // TripData o StopData
}

// Datos específicos de un viaje en movimiento
type TripData struct {
	DistanceKm  float64      `json:"distance_km"`
	MaxSpeed    int          `json:"max_speed"`
	AvgSpeed    int          `json:"avg_speed"`
	StartAddr   string       `json:"start_address,omitempty"` // Opcional (Reverse Geocoding)
	EndAddr     string       `json:"end_address,omitempty"`   // Opcional
	EncodedPath string       `json:"encoded_path,omitempty"`  // Opcional: Google Polyline Algorithm
	PathGeoJSON interface{}  `json:"path_geojson,omitempty"`  // Opcional: GeoJSON directo
	Points      []RoutePoint `json:"points,omitempty"`        // Lista detallada de puntos con metadata
	StartTime   time.Time    `json:"start_time"`
	EndTime     time.Time    `json:"end_time"`
	Duration    string       `json:"duration"`
}

// Datos de un punto individual en la ruta
type RoutePoint struct {
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Speed        int     `json:"speed"`
	BatteryLevel float64 `json:"battery_level"` // Nivel de batería (0-100 o voltaje)
	Course       int     `json:"course"`        // Dirección (0-360)
	Timestamp    string  `json:"timestamp"`     // Fecha y hora del punto
}

// Datos específicos de una parada
type StopData struct {
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	Address      string    `json:"address,omitempty"` // "Av. Banzer, 4to Anillo"
	BatteryLevel float64   `json:"battery_level"`     // Nivel de batería durante la parada
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Duration     string    `json:"duration"`
}

// Estructura temporal para recibir el resultado de la query
type RouteResult struct {
	Distance float64 `gorm:"column:dist"`
	GeoJSON  string  `gorm:"column:path"`
}
