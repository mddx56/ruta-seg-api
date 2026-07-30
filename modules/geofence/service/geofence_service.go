package service

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/geofence/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/geofence/repository"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type GeofenceService interface {
	Create(ctx context.Context, req dto.GeofenceCreateRequest, createdByID uuid.UUID) (dto.GeofenceResponse, error)
	Update(ctx context.Context, req dto.GeofenceUpdateRequest) (dto.GeofenceResponse, error)
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]dto.GeofenceResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (dto.GeofenceResponse, error)
}

type geofenceService struct {
	repo repository.GeofenceRepository
}

func NewGeofenceService(injector *do.Injector) (GeofenceService, error) {
	repo := do.MustInvoke[repository.GeofenceRepository](injector)
	return &geofenceService{
		repo: repo,
	}, nil
}

func (s *geofenceService) Create(ctx context.Context, req dto.GeofenceCreateRequest, createdByID uuid.UUID) (dto.GeofenceResponse, error) {
	geofence := entities.Geofence{
		Name:        req.Name,
		Type:        req.Type,
		Radius:      req.Radius,
		Points:      toPointEntities(uuid.Nil, req.Points),
		CreatedByID: createdByID,
	}

	if err := s.repo.Create(ctx, &geofence); err != nil {
		return dto.GeofenceResponse{}, err
	}

	return toResponse(geofence), nil
}

func (s *geofenceService) Update(ctx context.Context, req dto.GeofenceUpdateRequest) (dto.GeofenceResponse, error) {
	geofence, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return dto.GeofenceResponse{}, err
	}

	if req.Name != "" {
		geofence.Name = req.Name
	}
	if req.Type != "" {
		geofence.Type = req.Type
	}
	if req.Radius != nil {
		geofence.Radius = req.Radius
	}

	var points []entities.GeofencePoint
	if req.Points != nil {
		points = toPointEntities(geofence.ID, req.Points)
	}

	if err := s.repo.Update(ctx, &geofence, points); err != nil {
		return dto.GeofenceResponse{}, err
	}

	updated, err := s.repo.FindByID(ctx, geofence.ID)
	if err != nil {
		return dto.GeofenceResponse{}, err
	}

	return toResponse(updated), nil
}

func (s *geofenceService) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return s.repo.ChangeStatus(ctx, id)
}

func (s *geofenceService) FindAll(ctx context.Context) ([]dto.GeofenceResponse, error) {
	geofences, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.GeofenceResponse, 0, len(geofences))
	for _, g := range geofences {
		responses = append(responses, toResponse(g))
	}

	return responses, nil
}

func (s *geofenceService) FindByID(ctx context.Context, id uuid.UUID) (dto.GeofenceResponse, error) {
	geofence, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.GeofenceResponse{}, err
	}

	return toResponse(geofence), nil
}

func toPointEntities(geofenceID uuid.UUID, points []dto.GeofencePointRequest) []entities.GeofencePoint {
	result := make([]entities.GeofencePoint, 0, len(points))
	for _, p := range points {
		result = append(result, entities.GeofencePoint{
			GeofenceID: geofenceID,
			Latitude:   p.Latitude,
			Longitude:  p.Longitude,
			Sequence:   p.Sequence,
		})
	}
	return result
}

func toResponse(g entities.Geofence) dto.GeofenceResponse {
	points := make([]dto.GeofencePointResponse, 0, len(g.Points))
	for _, p := range g.Points {
		points = append(points, dto.GeofencePointResponse{
			ID:        p.ID,
			Latitude:  p.Latitude,
			Longitude: p.Longitude,
			Sequence:  p.Sequence,
		})
	}

	return dto.GeofenceResponse{
		ID:          g.ID,
		Name:        g.Name,
		Type:        g.Type,
		Radius:      g.Radius,
		Points:      points,
		CreatedByID: g.CreatedByID,
		CreatedAt:   g.CreatedAt,
		Status:      g.Status,
	}
}
