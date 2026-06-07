package service

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_incident/repository"
	"gorm.io/gorm"
)

type AlarmIncidentService interface {
}

type alarmIncidentService struct {
	alarmIncidentRepository repository.AlarmIncidentRepository
	db                            *gorm.DB
}

func NewAlarmIncidentService(
	alarmIncidentRepo repository.AlarmIncidentRepository,
	db *gorm.DB,
) AlarmIncidentService {
	return &alarmIncidentService{
		alarmIncidentRepository: alarmIncidentRepo,
		db:                            db,
	}
}
