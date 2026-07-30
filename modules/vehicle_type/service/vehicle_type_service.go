package service

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_type/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_type/repository"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type VehicleTypeService interface {
	Create(ctx context.Context, req dto.VehicleTypeCreateRequest) (dto.VehicleTypeResponse, error)
	Update(ctx context.Context, req dto.VehicleTypeUpdateRequest) (dto.VehicleTypeResponse, error)
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]dto.VehicleTypeResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (dto.VehicleTypeResponse, error)
	FindByCode(ctx context.Context, code string) (dto.VehicleTypeResponse, error)
}

type vehicleTypeService struct {
	repo repository.VehicleTypeRepository
}

func NewVehicleTypeService(injector *do.Injector) (VehicleTypeService, error) {
	repo := do.MustInvoke[repository.VehicleTypeRepository](injector)
	return &vehicleTypeService{
		repo: repo,
	}, nil
}

func (s *vehicleTypeService) Create(ctx context.Context, req dto.VehicleTypeCreateRequest) (dto.VehicleTypeResponse, error) {
	vehicleType := entities.VehicleType{
		TypeName: req.TypeName,
		Code:     req.Code,
	}

	if err := s.repo.Create(ctx, &vehicleType); err != nil {
		return dto.VehicleTypeResponse{}, err
	}

	return dto.VehicleTypeResponse{
		ID:       vehicleType.ID,
		TypeName: vehicleType.TypeName,
		Code:     vehicleType.Code,
	}, nil
}

func (s *vehicleTypeService) Update(ctx context.Context, req dto.VehicleTypeUpdateRequest) (dto.VehicleTypeResponse, error) {
	vehicleType, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return dto.VehicleTypeResponse{}, err
	}

	if req.TypeName != "" {
		vehicleType.TypeName = req.TypeName
	}

	if req.Code != nil {
		vehicleType.Code = req.Code
	}

	if err := s.repo.Update(ctx, &vehicleType); err != nil {
		return dto.VehicleTypeResponse{}, err
	}

	return dto.VehicleTypeResponse{
		ID:       vehicleType.ID,
		TypeName: vehicleType.TypeName,
		Code:     vehicleType.Code,
	}, nil
}

func (s  *vehicleTypeService) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return s.repo.ChangeStatus(ctx, id)
}

func (s *vehicleTypeService) FindAll(ctx context.Context) ([]dto.VehicleTypeResponse, error) {
	vehicleTypes, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var responses []dto.VehicleTypeResponse
	for _, vt := range vehicleTypes {
		responses = append(responses, dto.VehicleTypeResponse{
			ID:       vt.ID,
			TypeName: vt.TypeName,
			Code:     vt.Code,
		})
	}

	return responses, nil
}

func (s *vehicleTypeService) FindByID(ctx context.Context, id uuid.UUID) (dto.VehicleTypeResponse, error) {
	vehicleType, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.VehicleTypeResponse{}, err
	}

	return dto.VehicleTypeResponse{
		ID:       vehicleType.ID,
		TypeName: vehicleType.TypeName,
		Code:     vehicleType.Code,
	}, nil
}

func (s *vehicleTypeService) FindByCode(ctx context.Context, code string) (dto.VehicleTypeResponse, error) {
	vehicleType, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return dto.VehicleTypeResponse{}, err
	}

	return dto.VehicleTypeResponse{
		ID:       vehicleType.ID,
		TypeName: vehicleType.TypeName,
		Code:     vehicleType.Code,
	}, nil
}
