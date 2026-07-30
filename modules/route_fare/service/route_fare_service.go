package service

import (
	"context"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/route_fare/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/route_fare/repository"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type RouteFareService interface {
	Create(ctx context.Context, req dto.RouteFareCreateRequest) (dto.RouteFareResponse, error)
	Update(ctx context.Context, req dto.RouteFareUpdateRequest) (dto.RouteFareResponse, error)
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]dto.RouteFareResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (dto.RouteFareResponse, error)
}

type routeFareService struct {
	repo repository.RouteFareRepository
}

func NewRouteFareService(injector *do.Injector) (RouteFareService, error) {
	repo := do.MustInvoke[repository.RouteFareRepository](injector)
	return &routeFareService{
		repo: repo,
	}, nil
}

func (s *routeFareService) Create(ctx context.Context, req dto.RouteFareCreateRequest) (dto.RouteFareResponse, error) {
	effectiveFrom := time.Now()
	if req.EffectiveFrom != nil {
		effectiveFrom = *req.EffectiveFrom
	}

	fare := entities.RouteFare{
		RouteID:       req.RouteID,
		AmountPerLap:  req.AmountPerLap,
		EffectiveFrom: effectiveFrom,
	}

	if err := s.repo.Create(ctx, &fare); err != nil {
		return dto.RouteFareResponse{}, err
	}

	return toResponse(fare), nil
}

func (s *routeFareService) Update(ctx context.Context, req dto.RouteFareUpdateRequest) (dto.RouteFareResponse, error) {
	fare, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return dto.RouteFareResponse{}, err
	}

	if req.AmountPerLap != nil {
		fare.AmountPerLap = *req.AmountPerLap
	}
	if req.EffectiveFrom != nil {
		fare.EffectiveFrom = *req.EffectiveFrom
	}

	if err := s.repo.Update(ctx, &fare); err != nil {
		return dto.RouteFareResponse{}, err
	}

	return toResponse(fare), nil
}

func (s *routeFareService) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return s.repo.ChangeStatus(ctx, id)
}

func (s *routeFareService) FindAll(ctx context.Context) ([]dto.RouteFareResponse, error) {
	fares, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.RouteFareResponse, 0, len(fares))
	for _, f := range fares {
		responses = append(responses, toResponse(f))
	}

	return responses, nil
}

func (s *routeFareService) FindByID(ctx context.Context, id uuid.UUID) (dto.RouteFareResponse, error) {
	fare, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.RouteFareResponse{}, err
	}

	return toResponse(fare), nil
}

func toResponse(f entities.RouteFare) dto.RouteFareResponse {
	return dto.RouteFareResponse{
		ID:            f.ID,
		RouteID:       f.RouteID,
		AmountPerLap:  f.AmountPerLap,
		EffectiveFrom: f.EffectiveFrom,
		CreatedAt:     f.CreatedAt,
		Status:        f.Status,
	}
}
