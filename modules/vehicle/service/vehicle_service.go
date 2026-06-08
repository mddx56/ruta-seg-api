package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle/repository"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type VehicleService interface {
	Create(ctx context.Context, req dto.VehicleCreateRequest) (dto.VehicleResponse, error)
	Update(ctx context.Context, req dto.VehicleUpdateRequest) (dto.VehicleResponse, error)
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]dto.VehicleResponse, error)
	GetSimple(ctx context.Context, available bool) ([]dto.VehicleSimpleResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (dto.VehicleResponse, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]dto.VehicleResponse, error)
	FindByChassisFull(ctx context.Context, chassis string) (dto.VehicleFullResponse, error)
}

type vehicleService struct {
	repo repository.VehicleRepository
}

func NewVehicleService(injector *do.Injector) (VehicleService, error) {
	repo := do.MustInvoke[repository.VehicleRepository](injector)
	return &vehicleService{
		repo: repo,
	}, nil
}

func (s *vehicleService) Create(ctx context.Context, req dto.VehicleCreateRequest) (dto.VehicleResponse, error) {
	// Normalizar Placa
	req.Placa = strings.ToUpper(strings.TrimSpace(req.Placa))

	// Verificar si la placa ya existe
	if _, err := s.repo.FindByPlaca(ctx, req.Placa); err == nil {
		return dto.VehicleResponse{}, fmt.Errorf(dto.MESSAGE_FAILED_DUPLICATE_PLACA)
	}

	// Normalizar y verificar Chasis
	if req.Chassis != nil {
		normalizedChassis := strings.ToUpper(strings.TrimSpace(*req.Chassis))
		if normalizedChassis == "" {
			req.Chassis = nil
		} else {
			req.Chassis = &normalizedChassis
			// Verificar si el chasis ya existe
			if _, err := s.repo.FindByChassis(ctx, *req.Chassis); err == nil {
				return dto.VehicleResponse{}, fmt.Errorf(dto.MESSAGE_FAILED_DUPLICATE_CHASSIS)
			}
		}
	}

	vehicle := entities.Vehicle{
		Placa:       req.Placa,
		Description: req.Description,
		Year:        req.Year,
		KmLiter:     req.KmLiter,
		Chassis:     req.Chassis,
		Color:       req.Color,
		PhotoURL:    req.PhotoURL,
		UserID:      req.UserID,
		ModelID:     req.ModelID,
	}

	if err := s.repo.Create(ctx, &vehicle); err != nil {
		return dto.VehicleResponse{}, err
	}

	// Fetch again to get relations
	vehicle, err := s.repo.FindByID(ctx, vehicle.ID)
	if err != nil {
		return dto.VehicleResponse{}, err
	}

	return s.mapToResponse(vehicle), nil
}

func (s *vehicleService) Update(ctx context.Context, req dto.VehicleUpdateRequest) (dto.VehicleResponse, error) {
	vehicle, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return dto.VehicleResponse{}, err
	}

	if req.Placa != "" {
		// Verificar duplicado de placa al actualizar
		if existing, err := s.repo.FindByPlaca(ctx, req.Placa); err == nil {
			if existing.ID != req.ID {
				return dto.VehicleResponse{}, fmt.Errorf(dto.MESSAGE_FAILED_DUPLICATE_PLACA)
			}
		}
		vehicle.Placa = req.Placa
	}
	if req.Description != nil {
		vehicle.Description = req.Description
	}
	if req.Year != nil {
		vehicle.Year = req.Year
	}
	if req.KmLiter != nil {
		vehicle.KmLiter = req.KmLiter
	}
	if req.Chassis != nil {
		// Verificar duplicado de chasis al actualizar
		normalizedChassis := strings.ToUpper(strings.TrimSpace(*req.Chassis))
		if normalizedChassis == "" {
			req.Chassis = nil
		} else {
			req.Chassis = &normalizedChassis
			if existing, err := s.repo.FindByChassis(ctx, *req.Chassis); err == nil {
				if existing.ID != req.ID {
					return dto.VehicleResponse{}, fmt.Errorf(dto.MESSAGE_FAILED_DUPLICATE_CHASSIS)
				}
			}
		}
		vehicle.Chassis = req.Chassis
	}
	if req.Color != nil {
		vehicle.Color = req.Color
	}
	if req.PhotoURL != nil {
		vehicle.PhotoURL = req.PhotoURL
	}
	if req.UserID != uuid.Nil {
		vehicle.UserID = req.UserID
	}
	if req.ModelID != uuid.Nil {
		vehicle.ModelID = req.ModelID
	}

	if err := s.repo.Update(ctx, &vehicle); err != nil {
		return dto.VehicleResponse{}, err
	}

	// Fetch again to get updated relations
	vehicle, err = s.repo.FindByID(ctx, vehicle.ID)
	if err != nil {
		return dto.VehicleResponse{}, err
	}

	return s.mapToResponse(vehicle), nil
}

func (s  *vehicleService) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return s.repo.ChangeStatus(ctx, id)
}

func (s *vehicleService) FindAll(ctx context.Context) ([]dto.VehicleResponse, error) {
	vehicles, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.VehicleResponse, 0)
	for _, vehicle := range vehicles {
		responses = append(responses, s.mapToResponse(vehicle))
	}

	return responses, nil
}

func (s *vehicleService) GetSimple(ctx context.Context, available bool) ([]dto.VehicleSimpleResponse, error) {
	vehicles, err := s.repo.FindAllSimple(ctx, available)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.VehicleSimpleResponse, 0)
	for _, vehicle := range vehicles {
		modelName := ""
		makeName := ""

		if vehicle.Model != nil {
			modelName = vehicle.Model.ModelName
			if vehicle.Model.Make != nil {
				makeName = vehicle.Model.Make.MakeName
			}
		}

		responses = append(responses, dto.VehicleSimpleResponse{
			ID:        vehicle.ID,
			Placa:     vehicle.Placa,
			Chassis:   vehicle.Chassis,
			ModelName: modelName,
			MakeName:  makeName,
		})
	}

	return responses, nil
}

func (s *vehicleService) FindByUserID(ctx context.Context, userID uuid.UUID) ([]dto.VehicleResponse, error) {
	vehicles, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return []dto.VehicleResponse{}, err
	}

	responses := make([]dto.VehicleResponse, 0)
	for _, v := range vehicles {
		res := s.mapToResponse(v)
		if len(v.Installations) > 0 {
			inst := v.Installations[0]
			imei := inst.Imei
			res.ActiveInstallation = &dto.VehicleInstallationInfo{
				InstallationID: inst.InstallationID,
				DeviceIMEI:     imei,
				InstalledAt:    inst.InstalledAt,
			}
		}
		responses = append(responses, res)
	}
	return responses, nil
}

func (s *vehicleService) FindByID(ctx context.Context, id uuid.UUID) (dto.VehicleResponse, error) {
	vehicle, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.VehicleResponse{}, err
	}

	return s.mapToResponse(vehicle), nil
}

func (s *vehicleService) FindByChassisFull(ctx context.Context, chassis string) (dto.VehicleFullResponse, error) {
	vehicle, err := s.repo.FindByChassisFull(ctx, chassis)
	if err != nil {
		return dto.VehicleFullResponse{}, err
	}

	base := s.mapToResponse(vehicle)
	res := dto.VehicleFullResponse{
		Vehicle: base,
		AvailableForInstallation: true,
	}

	if vehicle.Model != nil && vehicle.Model.VehicleType != nil {
		res.VehicleType = &dto.VehicleTypeInfo{
			ID:   vehicle.Model.VehicleType.ID,
			Name: vehicle.Model.VehicleType.TypeName,
		}
	}

	// Active installation (if any)
	if len(vehicle.Installations) > 0 {
		res.AvailableForInstallation = false
		inst := vehicle.Installations[0]
		active := &dto.VehicleFullInstallationInfo{
			InstallationID: inst.InstallationID,
			InstalledAt:    inst.InstalledAt,
			InstallReason:  inst.InstallReason,
		}
		if inst.Device != nil {
			active.Device = &dto.VehicleFullDeviceInfo{
				IMEI:           inst.Device.IMEI,
				Model:          inst.Device.Model,
				SimPhoneNumber: inst.Device.SimPhoneNumber,
				SimProvider:    inst.Device.SimProvider,
			}
		}
		res.ActiveInstallation = active
	}

	return res, nil
}

func (s *vehicleService) mapToResponse(vehicle entities.Vehicle) dto.VehicleResponse {
	response := dto.VehicleResponse{
		ID:          vehicle.ID,
		Placa:       vehicle.Placa,
		Description: vehicle.Description,
		Year:        vehicle.Year,
		KmLiter:     vehicle.KmLiter,
		Chassis:     vehicle.Chassis,
		Color:       vehicle.Color,
		PhotoURL:    vehicle.PhotoURL,
		CreatedAt:   vehicle.CreatedAt,
		UpdatedAt:   vehicle.UpdatedAt,
		Status:      vehicle.Status,
	}

	if vehicle.User != nil {
		response.User = &dto.UserInfo{
			ID:    vehicle.User.ID,
			Name:  vehicle.User.Name,
			Email: vehicle.User.Email,
		}
	}

	if vehicle.Model != nil {
		modelInfo := &dto.ModelInfo{
			ID:        vehicle.Model.ID,
			ModelName: vehicle.Model.ModelName,
		}

		if vehicle.Model.Make != nil {
			modelInfo.Make = &dto.MakeInfo{
				ID:       vehicle.Model.Make.ID,
				MakeName: vehicle.Model.Make.MakeName,
			}
		}

		response.Model = modelInfo
	}

	return response
}
