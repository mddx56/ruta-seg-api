package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AlarmType - Catálogo de tipos de alarma (Exceso de Velocidad, Geocerca, etc.)
type AlarmType struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Code        string    `gorm:"type:varchar(50);unique;not null" json:"code"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Severity    string    `gorm:"type:varchar(20);default:'WARNING'" json:"severity"`
	Timestamp
}

func (a *AlarmType) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}

// AlarmRule - Programación de la alarma que configura el cliente
type AlarmRule struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	
	AlarmTypeID uuid.UUID `gorm:"type:uuid;not null" json:"alarm_type_id"`
	AlarmType   AlarmType `gorm:"foreignKey:AlarmTypeID" json:"alarm_type"`

	Devices     []Device   `gorm:"many2many:alarm_rule_devices;joinForeignKey:AlarmRuleID;joinReferences:DeviceIMEI" json:"devices,omitempty"` 
	SpeedLimit  *float64   `gorm:"type:decimal(10,2)" json:"speed_limit,omitempty"` 
	GeofenceID  *uuid.UUID `gorm:"type:uuid" json:"geofence_id,omitempty"`
	Geofence    *Geofence  `gorm:"foreignKey:GeofenceID" json:"geofence,omitempty"`

	TimeStart   *time.Time `gorm:"type:time" json:"time_start,omitempty"`
	TimeEnd     *time.Time `gorm:"type:time" json:"time_end,omitempty"`
	DaysOfWeek  int        `gorm:"not null;default:127" json:"days_of_week"` 

	IsActive    bool       `gorm:"default:true" json:"is_active"`
	
	CreatedByID uuid.UUID  `gorm:"type:uuid;not null" json:"created_by_id"`
	User        *User      `gorm:"foreignKey:CreatedByID" json:"user,omitempty"`
	Timestamp
}

func (a *AlarmRule) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}

// AlarmIncident - El registro de cuando el GPS viola la regla
type AlarmIncident struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	
	DeviceID    uuid.UUID `gorm:"type:uuid;not null" json:"device_id"`
	
	AlarmRuleID uuid.UUID `gorm:"type:uuid;not null" json:"alarm_rule_id"`
	AlarmRule   AlarmRule `gorm:"foreignKey:AlarmRuleID" json:"alarm_rule"`

	Latitude    float64   `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude   float64   `gorm:"type:decimal(11,8);not null" json:"longitude"`
	Speed       float64   `gorm:"type:decimal(10,2)" json:"speed"`
	EventTime   time.Time `gorm:"not null" json:"event_time"` 
	
	IsResolved  bool       `gorm:"default:false" json:"is_resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy  *uuid.UUID `gorm:"type:uuid" json:"resolved_by,omitempty"`
	Notes       string     `gorm:"type:text" json:"notes,omitempty"`

	Timestamp
}

func (a *AlarmIncident) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}
