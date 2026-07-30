package service

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/route/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/route/repository"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type RouteService interface {
	Create(ctx context.Context, req dto.RouteCreateRequest, createdByID uuid.UUID) (dto.RouteResponse, error)
	Update(ctx context.Context, req dto.RouteUpdateRequest) (dto.RouteResponse, error)
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]dto.RouteResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (dto.RouteResponse, error)
}

type routeService struct {
	repo repository.RouteRepository
}

func NewRouteService(injector *do.Injector) (RouteService, error) {
	repo := do.MustInvoke[repository.RouteRepository](injector)
	return &routeService{
		repo: repo,
	}, nil
}

func (s *routeService) Create(ctx context.Context, req dto.RouteCreateRequest, createdByID uuid.UUID) (dto.RouteResponse, error) {
	route := entities.Route{
		Name:                      req.Name,
		Description:               req.Description,
		MapColor:                  req.MapColor,
		CheckpointGeofenceID:      req.CheckpointGeofenceID,
		AllowedLapDurationSeconds: req.AllowedLapDurationSeconds,
		MaxSpeedKmh:               req.MaxSpeedKmh,
		MaxStopDurationSeconds:    req.MaxStopDurationSeconds,
		Geometry:                  req.Geometry,
		Active:                    true,
		CreatedByID:               createdByID,
		Stops:                     toStopEntities(uuid.Nil, req.Stops),
	}

	if err := s.repo.Create(ctx, &route); err != nil {
		return dto.RouteResponse{}, err
	}

	return toResponse(route), nil
}

func (s *routeService) Update(ctx context.Context, req dto.RouteUpdateRequest) (dto.RouteResponse, error) {
	route, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return dto.RouteResponse{}, err
	}

	if req.Name != "" {
		route.Name = req.Name
	}
	if req.Description != nil {
		route.Description = req.Description
	}
	if req.MapColor != nil {
		route.MapColor = req.MapColor
	}
	if req.CheckpointGeofenceID != nil {
		route.CheckpointGeofenceID = req.CheckpointGeofenceID
	}
	if req.AllowedLapDurationSeconds != nil {
		route.AllowedLapDurationSeconds = req.AllowedLapDurationSeconds
	}
	if req.MaxSpeedKmh != nil {
		route.MaxSpeedKmh = req.MaxSpeedKmh
	}
	if req.MaxStopDurationSeconds != nil {
		route.MaxStopDurationSeconds = req.MaxStopDurationSeconds
	}
	if req.Geometry != nil {
		route.Geometry = req.Geometry
	}
	if req.Active != nil {
		route.Active = *req.Active
	}

	var stops []entities.RouteStop
	if req.Stops != nil {
		stops = toStopEntities(route.ID, req.Stops)
	}

	if err := s.repo.Update(ctx, &route, stops); err != nil {
		return dto.RouteResponse{}, err
	}

	updated, err := s.repo.FindByID(ctx, route.ID)
	if err != nil {
		return dto.RouteResponse{}, err
	}

	return toResponse(updated), nil
}

func (s *routeService) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return s.repo.ChangeStatus(ctx, id)
}

func (s *routeService) FindAll(ctx context.Context) ([]dto.RouteResponse, error) {
	routes, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.RouteResponse, 0, len(routes))
	for _, r := range routes {
		responses = append(responses, toResponse(r))
	}

	return responses, nil
}

func (s *routeService) FindByID(ctx context.Context, id uuid.UUID) (dto.RouteResponse, error) {
	route, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.RouteResponse{}, err
	}

	return toResponse(route), nil
}

func toStopEntities(routeID uuid.UUID, stops []dto.RouteStopRequest) []entities.RouteStop {
	result := make([]entities.RouteStop, 0, len(stops))
	for _, s := range stops {
		result = append(result, entities.RouteStop{
			RouteID:   routeID,
			Name:      s.Name,
			Latitude:  s.Latitude,
			Longitude: s.Longitude,
			Sequence:  s.Sequence,
		})
	}
	return result
}

func toResponse(r entities.Route) dto.RouteResponse {
	stops := make([]dto.RouteStopResponse, 0, len(r.Stops))
	for _, s := range r.Stops {
		stops = append(stops, dto.RouteStopResponse{
			ID:        s.ID,
			Name:      s.Name,
			Latitude:  s.Latitude,
			Longitude: s.Longitude,
			Sequence:  s.Sequence,
		})
	}

	return dto.RouteResponse{
		ID:                        r.ID,
		Name:                      r.Name,
		Description:               r.Description,
		MapColor:                  r.MapColor,
		CheckpointGeofenceID:      r.CheckpointGeofenceID,
		AllowedLapDurationSeconds: r.AllowedLapDurationSeconds,
		MaxSpeedKmh:               r.MaxSpeedKmh,
		MaxStopDurationSeconds:    r.MaxStopDurationSeconds,
		Geometry:                  r.Geometry,
		Active:                    r.Active,
		Stops:                     stops,
		CreatedByID:               r.CreatedByID,
		CreatedAt:                 r.CreatedAt,
		Status:                    r.Status,
	}
}
