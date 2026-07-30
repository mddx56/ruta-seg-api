package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type VehicleRouteRepository interface {
	Create(ctx context.Context, vehicleRoute *entities.VehicleRoute) error
	Update(ctx context.Context, vehicleRoute *entities.VehicleRoute) error
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]entities.VehicleRoute, error)
	FindByID(ctx context.Context, id uuid.UUID) (entities.VehicleRoute, error)
}

type vehicleRouteRepository struct {
	db *gorm.DB
}

func NewVehicleRouteRepository(injector *do.Injector) (VehicleRouteRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &vehicleRouteRepository{
		db: db,
	}, nil
}

func (r *vehicleRouteRepository) Create(ctx context.Context, vehicleRoute *entities.VehicleRoute) error {
	return r.db.WithContext(ctx).Create(vehicleRoute).Error
}

func (r *vehicleRouteRepository) Update(ctx context.Context, vehicleRoute *entities.VehicleRoute) error {
	return r.db.WithContext(ctx).
		Model(&entities.VehicleRoute{}).
		Where("id = ?", vehicleRoute.ID).
		Select("route_id", "pin_number", "active", "updated_at").
		Updates(vehicleRoute).Error
}

func (r *vehicleRouteRepository) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entities.VehicleRoute{}).
		Where("id = ?", id).
		Update("status", gorm.Expr("NOT status")).Error
}

func (r *vehicleRouteRepository) FindAll(ctx context.Context) ([]entities.VehicleRoute, error) {
	var vehicleRoutes []entities.VehicleRoute
	err := r.db.WithContext(ctx).
		Where("status = ?", true).
		Order("pin_number ASC").
		Find(&vehicleRoutes).Error
	return vehicleRoutes, err
}

func (r *vehicleRouteRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.VehicleRoute, error) {
	var vehicleRoute entities.VehicleRoute
	err := r.db.WithContext(ctx).
		Where("status = ?", true).
		First(&vehicleRoute, "id = ?", id).Error
	return vehicleRoute, err
}
