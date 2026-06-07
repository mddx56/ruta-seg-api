package entities

import "time"

// DeviceLastPosition es una tabla de caché que mantiene SIEMPRE
// la posición más reciente de cada dispositivo.
//
// Se actualiza mediante UPSERT en cada INSERT de la tabla positions.
// Esto elimina completamente el costoso LATERAL JOIN en el dashboard,
// convirtiendo: O(N × P) → O(N) — constante sin importar el historial.
type DeviceLastPosition struct {
	IMEI       string    `gorm:"primaryKey;type:varchar(20);column:imei" json:"imei"`
	Latitude   float64   `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude  float64   `gorm:"type:decimal(11,8);not null" json:"longitude"`
	Speed      int       `json:"speed"`
	Course     int       `json:"course"`
	DeviceTime time.Time `gorm:"index" json:"device_time"`
	ServerTime time.Time `gorm:"index" json:"server_time"`
	Attributes *string   `gorm:"type:jsonb" json:"attributes,omitempty"`
	UpdatedAt  time.Time `gorm:"default:NOW()" json:"updated_at"`
}

func (DeviceLastPosition) TableName() string {
	return "device_last_positions"
}
