package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type VehicleTypeRepository interface {
	Create(ctx context.Context, vehicleType *entities.VehicleType) error
	Update(ctx context.Context, vehicleType *entities.VehicleType) error
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]entities.VehicleType, error)
	FindByID(ctx context.Context, id uuid.UUID) (entities.VehicleType, error)
}

type vehicleTypeRepository struct {
	db *gorm.DB
}

func NewVehicleTypeRepository(injector *do.Injector) (VehicleTypeRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &vehicleTypeRepository{
		db: db,
	}, nil
}

func (r *vehicleTypeRepository) Create(ctx context.Context, vehicleType *entities.VehicleType) error {
	return r.db.WithContext(ctx).Create(vehicleType).Error
}

func (r *vehicleTypeRepository) Update(ctx context.Context, vehicleType *entities.VehicleType) error {
	// Use Select to explicitly specify which fields to update
	// This prevents GORM from trying to update preloaded associations
	return r.db.WithContext(ctx).
		Model(vehicleType).
		Select("type_name", "updated_at").
		Updates(vehicleType).Error
}

func (r  *vehicleTypeRepository) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Exec("UPDATE " + r.db.Statement.Table + " SET status = NOT status WHERE id = ?", id).Error
}

func (r *vehicleTypeRepository) FindAll(ctx context.Context) ([]entities.VehicleType, error) {
	var vehicleTypes []entities.VehicleType
	err := r.db.WithContext(ctx).Where("status = ?", true).Find(&vehicleTypes).Error
	return vehicleTypes, err
}

func (r *vehicleTypeRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.VehicleType, error) {
	var vehicleType entities.VehicleType
	err := r.db.WithContext(ctx).Where("status = ?", true).First(&vehicleType, id).Error
	return vehicleType, err
}
