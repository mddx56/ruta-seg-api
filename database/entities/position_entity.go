package entities

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Position struct {
	ID uint64 `gorm:"primaryKey" json:"id"`
	// Imei replaced DeviceID (string) FK
	Imei string `gorm:"column:device_id;index;type:varchar(20);not null" json:"imei"`

	// 1. PARA EL FRONTEND (Lectura rápida y JSON)
	Latitude  float64 `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude float64 `gorm:"type:decimal(11,8);not null" json:"longitude"`

	// 2. PARA LA BASE DE DATOS (Cálculos espaciales potentes)
	// 'type:gist' crea un índice espacial ultra-rápido
	Geom string `gorm:"type:geometry(Point,4326);index:idx_positions_geom,type:gist" json:"-"`

	Speed      int       `json:"speed"`
	Course     int       `json:"course"`
	DeviceTime time.Time `gorm:"index" json:"device_time"`

	// Resto de campos
	ServerTime time.Time `gorm:"default:NOW()" json:"server_time"`
	Attributes *string   `gorm:"type:jsonb" json:"attributes,omitempty"`

	Device *Device `gorm:"foreignKey:Imei;references:IMEI" json:"device,omitempty"`
}

// 3. LA MAGIA: Hook BeforeCreate
// Antes de guardar en la BD, convierte Lat/Lon en un punto geométrico PostGIS
func (p *Position) BeforeCreate(tx *gorm.DB) (err error) {
	// Usamos formato EWKT que PostGIS entiende nativamente: "SRID=4326;POINT(lon lat)"
	// Esto evita problemas de interpolación con gorm.Expr y ST_MakePoint en el INSERT
	geomEWKT := fmt.Sprintf("SRID=4326;POINT(%v %v)", p.Longitude, p.Latitude)
	tx.Statement.SetColumn("Geom", geomEWKT)
	return
}

