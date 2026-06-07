package controller

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_incident/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_incident/validation"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type (
	AlarmIncidentController interface {
	}

	alarmIncidentController struct {
		alarmIncidentService    service.AlarmIncidentService
		alarmIncidentValidation *validation.AlarmIncidentValidation
		db                             *gorm.DB
	}
)

func NewAlarmIncidentController(injector *do.Injector, s service.AlarmIncidentService) AlarmIncidentController {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	alarmIncidentValidation := validation.NewAlarmIncidentValidation()
	return &alarmIncidentController{
		alarmIncidentService:    s,
		alarmIncidentValidation: alarmIncidentValidation,
		db:                             db,
	}
}
