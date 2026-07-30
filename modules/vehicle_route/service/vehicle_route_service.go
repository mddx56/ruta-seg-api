package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	modeldto "github.com/Caknoooo/go-gin-clean-starter/modules/model/dto"
	modelsvc "github.com/Caknoooo/go-gin-clean-starter/modules/model/service"
	vehicledto "github.com/Caknoooo/go-gin-clean-starter/modules/vehicle/dto"
	vehiclesvc "github.com/Caknoooo/go-gin-clean-starter/modules/vehicle/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_route/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_route/repository"
	vehicletypesvc "github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_type/service"
	"github.com/google/uuid"
	"github.com/samber/do"
)

// busVehicleTypeCode identifica el VehicleType de microbuses de forma estable,
// sin depender de su UUID (ver database/entities/vehicle_type_entity.go).
const busVehicleTypeCode = "BUS"

type VehicleRouteService interface {
	Create(ctx context.Context, req dto.VehicleRouteCreateRequest) (dto.VehicleRouteResponse, error)
	Update(ctx context.Context, req dto.VehicleRouteUpdateRequest) (dto.VehicleRouteResponse, error)
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]dto.VehicleRouteResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (dto.VehicleRouteResponse, error)
	RegisterMicro(ctx context.Context, req dto.RegisterMicroRequest) (dto.RegisterMicroResponse, error)
}

type vehicleRouteService struct {
	repo           repository.VehicleRouteRepository
	vehicleSvc     vehiclesvc.VehicleService
	modelSvc       modelsvc.ModelService
	vehicleTypeSvc vehicletypesvc.VehicleTypeService
}

func NewVehicleRouteService(injector *do.Injector) (VehicleRouteService, error) {
	repo := do.MustInvoke[repository.VehicleRouteRepository](injector)
	vehicleSvc := do.MustInvoke[vehiclesvc.VehicleService](injector)
	modelSvc := do.MustInvoke[modelsvc.ModelService](injector)
	vehicleTypeSvc := do.MustInvoke[vehicletypesvc.VehicleTypeService](injector)
	return &vehicleRouteService{
		repo:           repo,
		vehicleSvc:     vehicleSvc,
		modelSvc:       modelSvc,
		vehicleTypeSvc: vehicleTypeSvc,
	}, nil
}

func (s *vehicleRouteService) Create(ctx context.Context, req dto.VehicleRouteCreateRequest) (dto.VehicleRouteResponse, error) {
	assignedAt := time.Now()
	if req.AssignedAt != nil {
		assignedAt = *req.AssignedAt
	}

	vehicleRoute := entities.VehicleRoute{
		VehicleID:  req.VehicleID,
		RouteID:    req.RouteID,
		PinNumber:  req.PinNumber,
		Active:     true,
		AssignedAt: assignedAt,
	}

	if err := s.repo.Create(ctx, &vehicleRoute); err != nil {
		return dto.VehicleRouteResponse{}, err
	}

	return toResponse(vehicleRoute), nil
}

func (s *vehicleRouteService) RegisterMicro(ctx context.Context, req dto.RegisterMicroRequest) (dto.RegisterMicroResponse, error) {
	if req.RouteID != nil && (req.PinNumber == nil || *req.PinNumber == "") {
		return dto.RegisterMicroResponse{}, fmt.Errorf("pin_number es requerido cuando se especifica route_id")
	}

	modelID, err := s.resolveMicroModelID(ctx, req)
	if err != nil {
		return dto.RegisterMicroResponse{}, err
	}

	vehicle, err := s.vehicleSvc.Create(ctx, vehicledto.VehicleCreateRequest{
		Placa:       req.Placa,
		Description: req.Description,
		Year:        req.Year,
		KmLiter:     req.KmLiter,
		Chassis:     req.Chassis,
		Color:       req.Color,
		PhotoURL:    req.PhotoURL,
		UserID:      req.UserID,
		ModelID:     modelID,
	})
	if err != nil {
		return dto.RegisterMicroResponse{}, err
	}

	res := dto.RegisterMicroResponse{Vehicle: vehicle}

	if req.RouteID != nil {
		vehicleRoute, err := s.Create(ctx, dto.VehicleRouteCreateRequest{
			VehicleID: vehicle.ID,
			RouteID:   *req.RouteID,
			PinNumber: *req.PinNumber,
		})
		if err != nil {
			return dto.RegisterMicroResponse{}, err
		}
		res.VehicleRoute = &vehicleRoute
	}

	return res, nil
}

// resolveMicroModelID devuelve el ModelID a usar para el Vehicle: si el caller
// mandó model_id lo usa tal cual; si no, busca (o crea) un Model bajo el
// VehicleType "BUS" con el make_id/model_name recibidos.
func (s *vehicleRouteService) resolveMicroModelID(ctx context.Context, req dto.RegisterMicroRequest) (uuid.UUID, error) {
	if req.ModelID != nil {
		return *req.ModelID, nil
	}

	if req.MakeID == nil || req.ModelName == nil || *req.ModelName == "" {
		return uuid.Nil, fmt.Errorf("model_id, o make_id junto con model_name, son requeridos")
	}

	busType, err := s.vehicleTypeSvc.FindByCode(ctx, busVehicleTypeCode)
	if err != nil {
		return uuid.Nil, fmt.Errorf("no existe el vehicle_type con code=%q, créalo primero vía /api/vehicle-types: %w", busVehicleTypeCode, err)
	}

	if existingModel, err := s.modelSvc.FindByNameAndMake(ctx, *req.ModelName, *req.MakeID); err == nil {
		return existingModel.ID, nil
	}

	createdModel, err := s.modelSvc.Create(ctx, modeldto.ModelCreateRequest{
		Name:          *req.ModelName,
		VehicleTypeID: busType.ID,
		MakeID:        *req.MakeID,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return createdModel.ID, nil
}

func (s *vehicleRouteService) Update(ctx context.Context, req dto.VehicleRouteUpdateRequest) (dto.VehicleRouteResponse, error) {
	vehicleRoute, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return dto.VehicleRouteResponse{}, err
	}

	if req.RouteID != nil {
		vehicleRoute.RouteID = *req.RouteID
	}
	if req.PinNumber != "" {
		vehicleRoute.PinNumber = req.PinNumber
	}
	if req.Active != nil {
		vehicleRoute.Active = *req.Active
	}

	if err := s.repo.Update(ctx, &vehicleRoute); err != nil {
		return dto.VehicleRouteResponse{}, err
	}

	return toResponse(vehicleRoute), nil
}

func (s *vehicleRouteService) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return s.repo.ChangeStatus(ctx, id)
}

func (s *vehicleRouteService) FindAll(ctx context.Context) ([]dto.VehicleRouteResponse, error) {
	vehicleRoutes, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.VehicleRouteResponse, 0, len(vehicleRoutes))
	for _, vr := range vehicleRoutes {
		responses = append(responses, toResponse(vr))
	}

	return responses, nil
}

func (s *vehicleRouteService) FindByID(ctx context.Context, id uuid.UUID) (dto.VehicleRouteResponse, error) {
	vehicleRoute, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.VehicleRouteResponse{}, err
	}

	return toResponse(vehicleRoute), nil
}

func toResponse(vr entities.VehicleRoute) dto.VehicleRouteResponse {
	return dto.VehicleRouteResponse{
		ID:         vr.ID,
		VehicleID:  vr.VehicleID,
		RouteID:    vr.RouteID,
		PinNumber:  vr.PinNumber,
		Active:     vr.Active,
		AssignedAt: vr.AssignedAt,
		CreatedAt:  vr.CreatedAt,
		Status:     vr.Status,
	}
}
