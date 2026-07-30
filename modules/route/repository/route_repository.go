package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type RouteRepository interface {
	Create(ctx context.Context, route *entities.Route) error
	Update(ctx context.Context, route *entities.Route, stops []entities.RouteStop) error
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]entities.Route, error)
	FindByID(ctx context.Context, id uuid.UUID) (entities.Route, error)
}

type routeRepository struct {
	db *gorm.DB
}

func NewRouteRepository(injector *do.Injector) (RouteRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &routeRepository{
		db: db,
	}, nil
}

func (r *routeRepository) Create(ctx context.Context, route *entities.Route) error {
	return r.db.WithContext(ctx).Create(route).Error
}

// Update actualiza los campos base de la ruta y, si stops no es nil, reemplaza
// la lista completa de paradas (borra las existentes e inserta las nuevas).
func (r *routeRepository) Update(ctx context.Context, route *entities.Route, stops []entities.RouteStop) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entities.Route{}).
			Where("id = ?", route.ID).
			Select("name", "description", "map_color", "checkpoint_geofence_id", "allowed_lap_duration_seconds", "max_speed_kmh", "max_stop_duration_seconds", "geometry", "active", "updated_at").
			Updates(route).Error; err != nil {
			return err
		}

		if stops == nil {
			return nil
		}

		if err := tx.Where("route_id = ?", route.ID).Delete(&entities.RouteStop{}).Error; err != nil {
			return err
		}

		if len(stops) > 0 {
			if err := tx.Create(&stops).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *routeRepository) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entities.Route{}).
		Where("id = ?", id).
		Update("status", gorm.Expr("NOT status")).Error
}

func (r *routeRepository) FindAll(ctx context.Context) ([]entities.Route, error) {
	var routes []entities.Route
	err := r.db.WithContext(ctx).
		Preload("Stops", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence ASC")
		}).
		Preload("Geofence").
		Where("status = ?", true).
		Order("name ASC").
		Find(&routes).Error
	return routes, err
}

func (r *routeRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Route, error) {
	var route entities.Route
	err := r.db.WithContext(ctx).
		Preload("Stops", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence ASC")
		}).
		Preload("Geofence").
		Where("status = ?", true).
		First(&route, "id = ?", id).Error
	return route, err
}
