package service

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_rule/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_rule/repository"
	"gorm.io/gorm"
)

type AlarmRuleService interface {
	Create(ctx context.Context, req dto.AlarmRuleCreateRequest) (dto.AlarmRuleResponse, error)
	FindAll(ctx context.Context) ([]dto.AlarmRuleResponse, error)
}

type alarmRuleService struct {
	alarmRuleRepository repository.AlarmRuleRepository
	db                  *gorm.DB
}

func NewAlarmRuleService(
	alarmRuleRepo repository.AlarmRuleRepository,
	db *gorm.DB,
) AlarmRuleService {
	return &alarmRuleService{
		alarmRuleRepository: alarmRuleRepo,
		db:                  db,
	}
}

func (s *alarmRuleService) Create(ctx context.Context, req dto.AlarmRuleCreateRequest) (dto.AlarmRuleResponse, error) {
	days := req.DaysOfWeek
	if days == 0 {
		days = 127 // Todos los días por defecto si no se especifica
	}

	var devices []entities.Device
	for _, imei := range req.DeviceIMEIs {
		devices = append(devices, entities.Device{IMEI: imei})
	}

	alarmRule := entities.AlarmRule{
		Name:        req.Name,
		AlarmTypeID: req.AlarmTypeID,
		Devices:     devices,
		SpeedLimit:  req.SpeedLimit,
		GeofenceID:  req.GeofenceID,
		TimeStart:   req.TimeStart,
		TimeEnd:     req.TimeEnd,
		DaysOfWeek:  days,
		IsActive:    req.IsActive,
		CreatedByID: req.CreatedByID,
	}

	if err := s.alarmRuleRepository.Create(ctx, &alarmRule); err != nil {
		return dto.AlarmRuleResponse{}, err
	}

	return dto.AlarmRuleResponse{
		ID:          alarmRule.ID,
		Name:        alarmRule.Name,
		AlarmTypeID: alarmRule.AlarmTypeID,
		DeviceIMEIs: req.DeviceIMEIs,
		SpeedLimit:  alarmRule.SpeedLimit,
		GeofenceID:  alarmRule.GeofenceID,
		TimeStart:   alarmRule.TimeStart,
		TimeEnd:     alarmRule.TimeEnd,
		DaysOfWeek:  alarmRule.DaysOfWeek,
		IsActive:    alarmRule.IsActive,
		CreatedByID: alarmRule.CreatedByID,
	}, nil
}

func (s *alarmRuleService) FindAll(ctx context.Context) ([]dto.AlarmRuleResponse, error) {
	alarmRules, err := s.alarmRuleRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var responses []dto.AlarmRuleResponse
	for _, ar := range alarmRules {
		var imeis []string
		for _, dev := range ar.Devices {
			imeis = append(imeis, dev.IMEI)
		}
		responses = append(responses, dto.AlarmRuleResponse{
			ID:          ar.ID,
			Name:        ar.Name,
			AlarmTypeID: ar.AlarmTypeID,
			DeviceIMEIs: imeis,
			SpeedLimit:  ar.SpeedLimit,
			GeofenceID:  ar.GeofenceID,
			TimeStart:   ar.TimeStart,
			TimeEnd:     ar.TimeEnd,
			DaysOfWeek:  ar.DaysOfWeek,
			IsActive:    ar.IsActive,
			CreatedByID: ar.CreatedByID,
		})
	}
	return responses, nil
}
