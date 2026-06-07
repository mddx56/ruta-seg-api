package service

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/model/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/model/repository"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type ModelService interface {
	Create(ctx context.Context, req dto.ModelCreateRequest) (dto.ModelResponse, error)
	Update(ctx context.Context, req dto.ModelUpdateRequest) (dto.ModelResponse, error)
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]dto.ModelResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (dto.ModelResponse, error)
}

type modelService struct {
	repo repository.ModelRepository
}

func NewModelService(injector *do.Injector) (ModelService, error) {
	repo := do.MustInvoke[repository.ModelRepository](injector)
	return &modelService{
		repo: repo,
	}, nil
}

func (s *modelService) Create(ctx context.Context, req dto.ModelCreateRequest) (dto.ModelResponse, error) {
	model := entities.Model{
		ModelName:     req.Name,
		VehicleTypeID: req.VehicleTypeID,
		MakeID:        req.MakeID,
	}

	if err := s.repo.Create(ctx, &model); err != nil {
		return dto.ModelResponse{}, err
	}

	// Fetch again to get relations
	model, err := s.repo.FindByID(ctx, model.ID)
	if err != nil {
		return dto.ModelResponse{}, err
	}

	return s.mapToResponse(model), nil
}

func (s *modelService) Update(ctx context.Context, req dto.ModelUpdateRequest) (dto.ModelResponse, error) {
	model, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return dto.ModelResponse{}, err
	}

	if req.Name != "" {
		model.ModelName = req.Name
	}
	if req.VehicleTypeID != uuid.Nil {
		model.VehicleTypeID = req.VehicleTypeID
	}
	if req.MakeID != uuid.Nil {
		model.MakeID = req.MakeID
	}

	if err := s.repo.Update(ctx, &model); err != nil {
		return dto.ModelResponse{}, err
	}

	// Fetch again to get updated relations
	model, err = s.repo.FindByID(ctx, model.ID)
	if err != nil {
		return dto.ModelResponse{}, err
	}

	return s.mapToResponse(model), nil
}

func (s  *modelService) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return s.repo.ChangeStatus(ctx, id)
}

func (s *modelService) FindAll(ctx context.Context) ([]dto.ModelResponse, error) {
	models, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var responses []dto.ModelResponse
	for _, model := range models {
		responses = append(responses, s.mapToResponse(model))
	}

	return responses, nil
}

func (s *modelService) FindByID(ctx context.Context, id uuid.UUID) (dto.ModelResponse, error) {
	model, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.ModelResponse{}, err
	}

	return s.mapToResponse(model), nil
}

func (s *modelService) mapToResponse(model entities.Model) dto.ModelResponse {
	response := dto.ModelResponse{
		ID:        model.ID,
		Name:      model.ModelName,
		CreatedAt: model.CreatedAt,
		Status:    model.Status,
	}

	if model.VehicleType != nil {
		response.VehicleType = &dto.VehicleTypeInfo{
			ID:       model.VehicleType.ID,
			TypeName: model.VehicleType.TypeName,
		}
	}

	if model.Make != nil {
		response.Make = &dto.MakeInfo{
			ID:       model.Make.ID,
			MakeName: model.Make.MakeName,
		}
	}

	return response
}
