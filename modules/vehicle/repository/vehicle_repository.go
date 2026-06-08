package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type VehicleRepository interface {
	Create(ctx context.Context, vehicle *entities.Vehicle) error
	Update(ctx context.Context, vehicle *entities.Vehicle) error
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]entities.Vehicle, error)
	FindAllSimple(ctx context.Context, available bool) ([]entities.Vehicle, error)
	FindByID(ctx context.Context, id uuid.UUID) (entities.Vehicle, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Vehicle, error)
	FindByPlaca(ctx context.Context, placa string) (entities.Vehicle, error)
	FindByChassis(ctx context.Context, chassis string) (entities.Vehicle, error)
	FindByChassisFull(ctx context.Context, chassis string) (entities.Vehicle, error)
}

type vehicleRepository struct {
	db *gorm.DB
}

func NewVehicleRepository(injector *do.Injector) (VehicleRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &vehicleRepository{
		db: db,
	}, nil
}

func (r *vehicleRepository) Create(ctx context.Context, vehicle *entities.Vehicle) error {
	return r.db.WithContext(ctx).Create(vehicle).Error
}

func (r *vehicleRepository) Update(ctx context.Context, vehicle *entities.Vehicle) error {
	// Use Select to explicitly specify which fields to update
	// This prevents GORM from trying to update preloaded associations
	return r.db.WithContext(ctx).
		Model(vehicle).
		Select("placa", "description", "year", "km_liter", "chassis", "color", "photo_url", "user_id", "model_id", "updated_at").
		Updates(vehicle).Error
}

func (r  *vehicleRepository) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Exec("UPDATE " + r.db.Statement.Table + " SET status = NOT status WHERE id = ?", id).Error
}

func (r *vehicleRepository) FindAll(ctx context.Context) ([]entities.Vehicle, error) {
	var vehicles []entities.Vehicle
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Model").
		Preload("Model.Make").
		Where("status = ?", true).
		Find(&vehicles).Error
	return vehicles, err
}

func (r *vehicleRepository) FindAllSimple(ctx context.Context, available bool) ([]entities.Vehicle, error) {
	var vehicles []entities.Vehicle
	query := r.db.WithContext(ctx).
		Select("id", "placa", "chassis", "model_id").
		Preload("Model").
		Preload("Model.Make").
		Where("status = ?", true)

	if available {
		var occupiedVehicleIDs []uuid.UUID
		r.db.Table("device_installations").
			Where("removed_at IS NULL AND status = ?", true).
			Pluck("vehicle_id", &occupiedVehicleIDs)

		if len(occupiedVehicleIDs) > 0 {
			query = query.Where("id NOT IN ?", occupiedVehicleIDs)
		}
	}

	err := query.Find(&vehicles).Error
	return vehicles, err
}

func (r *vehicleRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Vehicle, error) {
	var vehicle entities.Vehicle
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Model").
		Preload("Model.Make").
		Where("status = ?", true).
		First(&vehicle, id).Error
	return vehicle, err
}

func (r *vehicleRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Vehicle, error) {
	var vehicles []entities.Vehicle
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Model").
		Preload("Model.Make").
		Preload("Installations", "removed_at IS NULL AND status = ?", true).
		Preload("Installations.Device").
		Where("user_id = ? AND status = ?", userID, true).
		Find(&vehicles).Error
	return vehicles, err
}

func (r *vehicleRepository) FindByPlaca(ctx context.Context, placa string) (entities.Vehicle, error) {
	var vehicle entities.Vehicle
	// Check against all records regardless of status to avoid unique constraint violations
	err := r.db.WithContext(ctx).Where("placa = ?", placa).First(&vehicle).Error
	return vehicle, err
}

func (r *vehicleRepository) FindByChassis(ctx context.Context, chassis string) (entities.Vehicle, error) {
	var vehicle entities.Vehicle
	// Check against all records regardless of status to avoid unique constraint violations
	err := r.db.WithContext(ctx).Where("chassis = ?", chassis).First(&vehicle).Error
	return vehicle, err
}

func (r *vehicleRepository) FindByChassisFull(ctx context.Context, chassis string) (entities.Vehicle, error) {
	var vehicle entities.Vehicle
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Model").
		Preload("Model.Make").
		Preload("Model.VehicleType").
		Preload("Installations", "removed_at IS NULL AND status = ?", true).
		Preload("Installations.Device").
		Where("status = ?", true).
		Where("chassis = ?", chassis).
		First(&vehicle).Error
	return vehicle, err
}
