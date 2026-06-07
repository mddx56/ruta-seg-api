package dto

import (
	"time"
	"github.com/google/uuid"
)

type (
	AlarmRuleCreateRequest struct {
		Name        string     `json:"name" binding:"required"`
		AlarmTypeID uuid.UUID  `json:"alarm_type_id" binding:"required"`
		DeviceIMEIs []string   `json:"device_imeis"`
		SpeedLimit  *float64   `json:"speed_limit"`
		GeofenceID  *uuid.UUID `json:"geofence_id"`
		TimeStart   *time.Time `json:"time_start"`
		TimeEnd     *time.Time `json:"time_end"`
		DaysOfWeek  int        `json:"days_of_week"` 
		IsActive    bool       `json:"is_active"`
		CreatedByID uuid.UUID  `json:"created_by_id" binding:"required"`
	}

	AlarmRuleResponse struct {
		ID          uuid.UUID  `json:"id"`
		Name        string     `json:"name"`
		AlarmTypeID uuid.UUID  `json:"alarm_type_id"`
		DeviceIMEIs []string   `json:"device_imeis"`
		SpeedLimit  *float64   `json:"speed_limit"`
		GeofenceID  *uuid.UUID `json:"geofence_id"`
		TimeStart   *time.Time `json:"time_start"`
		TimeEnd     *time.Time `json:"time_end"`
		DaysOfWeek  int        `json:"days_of_week"`
		IsActive    bool       `json:"is_active"`
		CreatedByID uuid.UUID  `json:"created_by_id"`
	}
)
