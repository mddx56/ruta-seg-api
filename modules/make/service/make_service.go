package service

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/make/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/make/repository"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type MakeService interface {
	Create(ctx context.Context, req dto.MakeCreateRequest) (dto.MakeResponse, error)
	Update(ctx context.Context, req dto.MakeUpdateRequest) (dto.MakeResponse, error)
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]dto.MakeResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (dto.MakeResponse, error)
}

type makeService struct {
	repo repository.MakeRepository
}

func NewMakeService(injector *do.Injector) (MakeService, error) {
	repo := do.MustInvoke[repository.MakeRepository](injector)
	return &makeService{
		repo: repo,
	}, nil
}

func (s *makeService) Create(ctx context.Context, req dto.MakeCreateRequest) (dto.MakeResponse, error) {
	make := entities.Make{
		MakeName: req.Name,
	}

	if err := s.repo.Create(ctx, &make); err != nil {
		return dto.MakeResponse{}, err
	}

	return dto.MakeResponse{
		ID:   make.ID,
		Name: make.MakeName,
		// CreatedAt: make.CreatedAt,
		// Status:    make.Status,
	}, nil
}

func (s *makeService) Update(ctx context.Context, req dto.MakeUpdateRequest) (dto.MakeResponse, error) {
	make, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return dto.MakeResponse{}, err
	}

	if req.Name != "" {
		make.MakeName = req.Name
	}

	if err := s.repo.Update(ctx, &make); err != nil {
		return dto.MakeResponse{}, err
	}

	return dto.MakeResponse{
		ID:   make.ID,
		Name: make.MakeName,
		// CreatedAt: make.CreatedAt,
		// Status:    make.Status,
	}, nil
}

func (s  *makeService) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return s.repo.ChangeStatus(ctx, id)
}

func (s *makeService) FindAll(ctx context.Context) ([]dto.MakeResponse, error) {
	makes, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var responses []dto.MakeResponse
	for _, make := range makes {
		responses = append(responses, dto.MakeResponse{
			ID:   make.ID,
			Name: make.MakeName,
			// CreatedAt: make.CreatedAt,
			// Status:    make.Status,
		})
	}

	return responses, nil
}

func (s *makeService) FindByID(ctx context.Context, id uuid.UUID) (dto.MakeResponse, error) {
	make, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.MakeResponse{}, err
	}

	return dto.MakeResponse{
		ID:   make.ID,
		Name: make.MakeName,
		// CreatedAt: make.CreatedAt,
		// Status:    make.Status,
	}, nil
}
