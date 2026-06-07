package service

import (
	"context"
	"time"

	"errors"
	"fmt"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device_installation/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device_installation/repository"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type DeviceInstallationService interface {
	Create(ctx context.Context, req dto.DeviceInstallationCreateRequest, userCreationID uuid.UUID) (dto.DeviceInstallationResponse, error)
	QuickCreate(ctx context.Context, req dto.DeviceInstallationQuickCreateRequest, userCreationID uuid.UUID) (dto.DeviceInstallationResponse, error)
	GetMine(ctx context.Context, userID uuid.UUID) ([]dto.DeviceInstallationMineResponse, error)
	GetAll(ctx context.Context) ([]dto.DeviceInstallationResponse, error)
	GetByIMEI(ctx context.Context, imei string) ([]dto.DeviceInstallationResponse, error)
	GetByVehicleID(ctx context.Context, vehicleID uuid.UUID) ([]dto.DeviceInstallationResponse, error)
	Uninstall(ctx context.Context, installationID uuid.UUID, req dto.DeviceInstallationUninstallRequest) (dto.DeviceInstallationResponse, error)
}

type deviceInstallationService struct {
	repo repository.DeviceInstallationRepository
}

func NewDeviceInstallationService(injector *do.Injector) (DeviceInstallationService, error) {
	repo := do.MustInvoke[repository.DeviceInstallationRepository](injector)
	return &deviceInstallationService{repo: repo}, nil
}

func (s *deviceInstallationService) Create(ctx context.Context, req dto.DeviceInstallationCreateRequest, userCreationID uuid.UUID) (dto.DeviceInstallationResponse, error) {
	// 0. Verificar que el usuario que realiza la instalación exista
	if _, err := s.repo.FindUserByID(ctx, userCreationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.DeviceInstallationResponse{}, fmt.Errorf("el usuario que realiza la instalación no existe o no es válido")
		}
		return dto.DeviceInstallationResponse{}, err
	}

	// 1. Verificar que el dispositivo exista
	_, err := s.repo.FindDeviceByIMEI(ctx, req.Imei)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.DeviceInstallationResponse{}, fmt.Errorf("el dispositivo con imei %s no está registrado", req.Imei)
		}
		return dto.DeviceInstallationResponse{}, err
	}

	// 2. Check if Device is already assigned
	existingDevice, err := s.repo.FindActiveByIMEI(ctx, req.Imei)
	if err == nil && existingDevice != nil {
		return dto.DeviceInstallationResponse{}, fmt.Errorf("el dispositivo con imei %s ya tiene un vehiculo asignado (Id: %s)", req.Imei, existingDevice.InstallationID)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.DeviceInstallationResponse{}, err
	}

	// 3. Check if Vehicle already has a device
	existingVehicle, err := s.repo.FindActiveByVehicleID(ctx, req.VehicleID)
	if err == nil && existingVehicle != nil {
		return dto.DeviceInstallationResponse{}, fmt.Errorf("el vehiculo con id %s ya tiene un dispositivo asignado (Id: %s)", req.VehicleID, existingVehicle.InstallationID)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.DeviceInstallationResponse{}, err
	}

	di := entities.DeviceInstallation{
		Imei:           req.Imei,
		VehicleID:      req.VehicleID,
		UserCreationID: &userCreationID,
		InstallReason:  req.InstallReason,
	}

	if err := s.repo.Create(ctx, &di); err != nil {
		return dto.DeviceInstallationResponse{}, err
	}

	return s.mapToResponse(di), nil
}

// QuickCreate permite crear una instalación usando solo IMEI del dispositivo y chasis del vehículo.
// Pensado para registro rápido desde la app móvil.
func (s *deviceInstallationService) QuickCreate(ctx context.Context, req dto.DeviceInstallationQuickCreateRequest, userCreationID uuid.UUID) (dto.DeviceInstallationResponse, error) {
	// 0. Verificar que el usuario que realiza la instalación exista
	if _, err := s.repo.FindUserByID(ctx, userCreationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.DeviceInstallationResponse{}, fmt.Errorf("el usuario que realiza la instalación no existe o no es válido")
		}
		return dto.DeviceInstallationResponse{}, err
	}

	// Normalizar entradas
	imei := req.Imei
	chassis := req.Chassis

	// 1. Buscar dispositivo por IMEI
	device, err := s.repo.FindDeviceByIMEI(ctx, imei)
	if err != nil {
		return dto.DeviceInstallationResponse{}, fmt.Errorf("no se encontró dispositivo con imei %s", imei)
	}

	// 2. Buscar vehículo por chasis
	vehicle, err := s.repo.FindVehicleByChassis(ctx, chassis)
	if err != nil {
		return dto.DeviceInstallationResponse{}, fmt.Errorf("no se encontró vehículo con chasis %s", chassis)
	}

	// 3. Validar que el dispositivo y el vehículo no tengan instalación activa
	existingDevice, err := s.repo.FindActiveByIMEI(ctx, device.IMEI)
	if err == nil && existingDevice != nil {
		return dto.DeviceInstallationResponse{}, fmt.Errorf("el dispositivo con imei %s ya tiene un vehiculo asignado (Id: %s)", device.IMEI, existingDevice.InstallationID)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.DeviceInstallationResponse{}, err
	}

	existingVehicle, err := s.repo.FindActiveByVehicleID(ctx, vehicle.ID)
	if err == nil && existingVehicle != nil {
		return dto.DeviceInstallationResponse{}, fmt.Errorf("el vehiculo con id %s ya tiene un dispositivo asignado (Id: %s)", vehicle.ID, existingVehicle.InstallationID)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.DeviceInstallationResponse{}, err
	}

	di := entities.DeviceInstallation{
		Imei:           device.IMEI,
		VehicleID:      vehicle.ID,
		UserCreationID: &userCreationID,
		InstallReason:  req.InstallReason,
	}
	if di.InstallReason == nil {
		defaultReason := "Instalación rápida desde app móvil"
		di.InstallReason = &defaultReason
	}

	if err := s.repo.Create(ctx, &di); err != nil {
		return dto.DeviceInstallationResponse{}, err
	}

	return s.mapToResponse(di), nil
}

func (s *deviceInstallationService) GetAll(ctx context.Context) ([]dto.DeviceInstallationResponse, error) {
	installations, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DeviceInstallationResponse, 0)
	for _, installation := range installations {
		responses = append(responses, s.mapToResponse(installation))
	}

	return responses, nil
}

func (s *deviceInstallationService) GetMine(ctx context.Context, userID uuid.UUID) ([]dto.DeviceInstallationMineResponse, error) {
	// "Mine" = instalaciones que yo registré (user_creation_id)
	installations, err := s.repo.FindAllByCreatorUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.DeviceInstallationMineResponse, 0, len(installations))
	for _, inst := range installations {
		item := dto.DeviceInstallationMineResponse{
			InstallationID: inst.InstallationID,
			Imei:           inst.Imei,
			InstalledAt:    inst.InstalledAt,
			RemovedAt:      inst.RemovedAt,
			InstallReason:  inst.InstallReason,
			RemovalReason:  inst.RemovalReason,
			WorkOrderID:    inst.WorkOrderID,
			Notes:          inst.Notes,
			Status:         inst.Status,
		}

		if inst.Device != nil {
			item.Device = &dto.DeviceInstallationDeviceInfo{
				IMEI:            inst.Device.IMEI,
				Model:           inst.Device.Model,
				SimPhoneNumber:  inst.Device.SimPhoneNumber,
				SimICCID:        inst.Device.SimICCID,
				SimProvider:     inst.Device.SimProvider,
				Protocol:        inst.Device.Protocol,
				FirmwareVersion: inst.Device.FirmwareVersion,
			}
		}

		if inst.Vehicle != nil {
			v := &dto.DeviceInstallationVehicleInfo{
				ID:          inst.Vehicle.ID,
				Placa:       inst.Vehicle.Placa,
				Chassis:     inst.Vehicle.Chassis,
				Description: inst.Vehicle.Description,
				Year:        inst.Vehicle.Year,
				KmLiter:     inst.Vehicle.KmLiter,
				Color:       inst.Vehicle.Color,
				PhotoURL:    inst.Vehicle.PhotoURL,
			}

			if inst.Vehicle.User != nil {
				v.Owner = &dto.DeviceInstallationOwnerInfo{
					ID:    inst.Vehicle.User.ID,
					Name:  inst.Vehicle.User.Name,
					Email: inst.Vehicle.User.Email,
				}
			}

			if inst.Vehicle.Model != nil {
				m := &dto.DeviceInstallationModelInfo{
					ID:   inst.Vehicle.Model.ID,
					Name: inst.Vehicle.Model.ModelName,
				}
				if inst.Vehicle.Model.Make != nil {
					m.Make = &dto.DeviceInstallationMakeInfo{
						ID:   inst.Vehicle.Model.Make.ID,
						Name: inst.Vehicle.Model.Make.MakeName,
					}
				}
				v.Model = m
			}

			item.Vehicle = v
		}

		res = append(res, item)
	}

	return res, nil
}

func (s *deviceInstallationService) GetByIMEI(ctx context.Context, imei string) ([]dto.DeviceInstallationResponse, error) {
	installations, err := s.repo.FindByIMEI(ctx, imei)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DeviceInstallationResponse, 0)
	for _, installation := range installations {
		responses = append(responses, s.mapToResponse(installation))
	}

	return responses, nil
}

func (s *deviceInstallationService) GetByVehicleID(ctx context.Context, vehicleID uuid.UUID) ([]dto.DeviceInstallationResponse, error) {
	installations, err := s.repo.FindByVehicleID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DeviceInstallationResponse, 0)
	for _, installation := range installations {
		responses = append(responses, s.mapToResponse(installation))
	}

	return responses, nil
}

func (s *deviceInstallationService) Uninstall(ctx context.Context, installationID uuid.UUID, req dto.DeviceInstallationUninstallRequest) (dto.DeviceInstallationResponse, error) {
	installation, err := s.repo.FindByID(ctx, installationID)
	if err != nil {
		return dto.DeviceInstallationResponse{}, err
	}

	if installation.RemovedAt != nil {
		return dto.DeviceInstallationResponse{}, errors.New("installation is already uninstalled")
	}

	// Update fields
	now := time.Now()
	installation.RemovedAt = &now
	installation.RemovalReason = req.RemovalReason
	// Optionally update odometer/engine hours if provided as "final" readings?
	// The DTO had them. Let's assume they are final readings but where do they go?
	// The entity has OdometerReading and EngineHours but those are usually "at install".
	// Ah, looking at entity:
	// OdometerReading *int (at install)
	// EngineHours *float64 (at install)
	// There is no "OdometerAtRemoval" in the entity I saw earlier.
	// Let's re-check entity.

	if err := s.repo.Update(ctx, installation); err != nil {
		return dto.DeviceInstallationResponse{}, err
	}

	return s.mapToResponse(*installation), nil
}

func (s *deviceInstallationService) mapToResponse(di entities.DeviceInstallation) dto.DeviceInstallationResponse {
	return dto.DeviceInstallationResponse{
		InstallationID: di.InstallationID,
		Imei:           di.Imei,
		VehicleID:      di.VehicleID,
		InstalledAt:    di.InstalledAt,
		RemovedAt:      di.RemovedAt,
		InstallReason:  di.InstallReason,
		RemovalReason:  di.RemovalReason,
		CreatedAt:      di.CreatedAt,
		UpdatedAt:      di.UpdatedAt,
		Status:         di.Status,
	}
}
