package repository

import (
	"gorm.io/gorm"
)

type AlarmIncidentRepository interface {
}

type alarmIncidentRepository struct {
	db *gorm.DB
}

func NewAlarmIncidentRepository(db *gorm.DB) AlarmIncidentRepository {
	return &alarmIncidentRepository{
		db: db,
	}
}
