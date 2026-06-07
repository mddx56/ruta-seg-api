package service

import (
	"context"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/log_socket/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/log_socket/repository"
	"gorm.io/gorm"
)

type LogSocketService interface {
	Create(ctx context.Context, req dto.LogSocketCreateRequest) (dto.LogSocketResponse, error)
}

type logSocketService struct {
	logSocketRepository repository.LogSocketRepository
	db                  *gorm.DB
}

func NewLogSocketService(
	logSocketRepo repository.LogSocketRepository,
	db *gorm.DB,
) LogSocketService {
	return &logSocketService{
		logSocketRepository: logSocketRepo,
		db:                  db,
	}
}

func (s *logSocketService) Create(ctx context.Context, req dto.LogSocketCreateRequest) (dto.LogSocketResponse, error) {
	now := time.Now()
	logSocket := entities.LogSocket{
		Payload:   req.Payload,
		CreatedAt: now,
	}

	createdLog, err := s.logSocketRepository.Create(ctx, s.db, logSocket)
	if err != nil {
		return dto.LogSocketResponse{}, err
	}

	return dto.LogSocketResponse{
		LogID:     createdLog.LogID,
		Payload:   createdLog.Payload,
		CreatedAt: createdLog.CreatedAt,
	}, nil
}
