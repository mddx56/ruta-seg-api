package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type DeviceInstallationRepository interface {
	Create(ctx context.Context, di *entities.DeviceInstallation) error
	Update(ctx context.Context, di *entities.DeviceInstallation) error
	FindAll(ctx context.Context) ([]entities.DeviceInstallation, error)
	FindByIMEI(ctx context.Context, imei string) ([]entities.DeviceInstallation, error)
	FindByVehicleID(ctx context.Context, vehicleID uuid.UUID) ([]entities.DeviceInstallation, error)
	FindActiveByIMEI(ctx context.Context, imei string) (*entities.DeviceInstallation, error)
	FindActiveByVehicleID(ctx context.Context, vehicleID uuid.UUID) (*entities.DeviceInstallation, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entities.DeviceInstallation, error)
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]entities.DeviceInstallation, error)
	FindAllByCreatorUserID(ctx context.Context, userID uuid.UUID) ([]entities.DeviceInstallation, error)
	FindDeviceByIMEI(ctx context.Context, imei string) (*entities.Device, error)
	FindVehicleByChassis(ctx context.Context, chassis string) (*entities.Vehicle, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
}

type deviceInstallationRepository struct {
	db *gorm.DB
}

func NewDeviceInstallationRepository(injector *do.Injector) (DeviceInstallationRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &deviceInstallationRepository{db: db}, nil
}

func (r *deviceInstallationRepository) Create(ctx context.Context, di *entities.DeviceInstallation) error {
	return r.db.WithContext(ctx).Create(di).Error
}

func (r *deviceInstallationRepository) Update(ctx context.Context, di *entities.DeviceInstallation) error {
	return r.db.WithContext(ctx).Save(di).Error
}

func (r *deviceInstallationRepository) FindAll(ctx context.Context) ([]entities.DeviceInstallation, error) {
	var installations []entities.DeviceInstallation
	err := r.db.WithContext(ctx).
		Preload("Device").
		Preload("Vehicle").
		Find(&installations).Error
	return installations, err
}

func (r *deviceInstallationRepository) FindByIMEI(ctx context.Context, imei string) ([]entities.DeviceInstallation, error) {
	var installations []entities.DeviceInstallation
	err := r.db.WithContext(ctx).
		Where("imei = ?", imei).
		Preload("Vehicle").
		Find(&installations).Error
	return installations, err
}

func (r *deviceInstallationRepository) FindByVehicleID(ctx context.Context, vehicleID uuid.UUID) ([]entities.DeviceInstallation, error) {
	var installations []entities.DeviceInstallation
	err := r.db.WithContext(ctx).
		Where("vehicle_id = ?", vehicleID).
		Preload("Device").
		Find(&installations).Error
	return installations, err
}

func (r *deviceInstallationRepository) FindActiveByIMEI(ctx context.Context, imei string) (*entities.DeviceInstallation, error) {
	var installation entities.DeviceInstallation
	err := r.db.WithContext(ctx).
		Where("imei = ? AND removed_at IS NULL AND status = true", imei).
		Preload("Vehicle").
		First(&installation).Error
	if err != nil {
		return nil, err
	}
	return &installation, nil
}

func (r *deviceInstallationRepository) FindActiveByVehicleID(ctx context.Context, vehicleID uuid.UUID) (*entities.DeviceInstallation, error) {
	var installation entities.DeviceInstallation
	err := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND removed_at IS NULL AND status = true", vehicleID).
		Preload("Device").
		First(&installation).Error
	if err != nil {
		return nil, err
	}
	return &installation, nil
}

func (r *deviceInstallationRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.DeviceInstallation, error) {
	var installation entities.DeviceInstallation
	err := r.db.WithContext(ctx).
		Where("installation_id = ?", id).
		Preload("Device").
		Preload("Vehicle").
		First(&installation).Error
	if err != nil {
		return nil, err
	}
	return &installation, nil
}

func (r *deviceInstallationRepository) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]entities.DeviceInstallation, error) {
	var installations []entities.DeviceInstallation
	err := r.db.WithContext(ctx).
		Joins("JOIN vehicles ON vehicles.id = device_installations.vehicle_id").
		Where("vehicles.user_id = ?", userID).
		Where("device_installations.status = ?", true).
		Where("device_installations.removed_at IS NULL").
		Preload("Device").
		Preload("Vehicle").
		Preload("Vehicle.User").
		Preload("Vehicle.Model").
		Preload("Vehicle.Model.Make").
		Find(&installations).Error
	return installations, err
}

// FindAllByCreatorUserID retorna instalaciones activas creadas por un usuario (installer/admin).
func (r *deviceInstallationRepository) FindAllByCreatorUserID(ctx context.Context, userID uuid.UUID) ([]entities.DeviceInstallation, error) {
	var installations []entities.DeviceInstallation
	err := r.db.WithContext(ctx).
		Where("user_creation_id = ?", userID).
		Where("device_installations.status = ?", true).
		Where("device_installations.removed_at IS NULL").
		Order("installed_at desc").
		Preload("Device").
		Preload("Vehicle").
		Preload("Vehicle.User").
		Preload("Vehicle.Model").
		Preload("Vehicle.Model.Make").
		Find(&installations).Error
	return installations, err
}

func (r *deviceInstallationRepository) FindDeviceByIMEI(ctx context.Context, imei string) (*entities.Device, error) {
	var device entities.Device
	err := r.db.WithContext(ctx).Where("imei = ?", imei).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceInstallationRepository) FindVehicleByChassis(ctx context.Context, chassis string) (*entities.Vehicle, error) {
	var vehicle entities.Vehicle
	err := r.db.WithContext(ctx).Where("chassis = ?", chassis).First(&vehicle).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func (r *deviceInstallationRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
