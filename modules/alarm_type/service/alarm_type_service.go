package service

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_type/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_type/repository"
	"gorm.io/gorm"
)

type AlarmTypeService interface {
	Create(ctx context.Context, req dto.AlarmTypeCreateRequest) (dto.AlarmTypeResponse, error)
	FindAll(ctx context.Context) ([]dto.AlarmTypeResponse, error)
}

type alarmTypeService struct {
	alarmTypeRepository repository.AlarmTypeRepository
	db                  *gorm.DB
}

func NewAlarmTypeService(
	alarmTypeRepo repository.AlarmTypeRepository,
	db *gorm.DB,
) AlarmTypeService {
	return &alarmTypeService{
		alarmTypeRepository: alarmTypeRepo,
		db:                  db,
	}
}

func (s *alarmTypeService) Create(ctx context.Context, req dto.AlarmTypeCreateRequest) (dto.AlarmTypeResponse, error) {
	alarmType := entities.AlarmType{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Severity:    req.Severity,
	}

	if err := s.alarmTypeRepository.Create(ctx, &alarmType); err != nil {
		return dto.AlarmTypeResponse{}, err
	}

	return dto.AlarmTypeResponse{
		ID:          alarmType.ID,
		Code:        alarmType.Code,
		Name:        alarmType.Name,
		Description: alarmType.Description,
		Severity:    alarmType.Severity,
	}, nil
}

func (s *alarmTypeService) FindAll(ctx context.Context) ([]dto.AlarmTypeResponse, error) {
	alarmTypes, err := s.alarmTypeRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var responses []dto.AlarmTypeResponse
	for _, at := range alarmTypes {
		responses = append(responses, dto.AlarmTypeResponse{
			ID:          at.ID,
			Code:        at.Code,
			Name:        at.Name,
			Description: at.Description,
			Severity:    at.Severity,
		})
	}
	return responses, nil
}
