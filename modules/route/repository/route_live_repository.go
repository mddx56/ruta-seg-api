package repository

import (
	"context"
	"errors"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

// ActiveRouteVehicle es un micro asignado activamente a una ruta, junto con el IMEI
// de su dispositivo GPS actualmente instalado (para buscar su posición en Redis).
type ActiveRouteVehicle struct {
	VehicleID uuid.UUID
	PinNumber string
	IMEI      string
}

type RouteLiveRepository interface {
	FindActiveVehiclesForRoute(ctx context.Context, routeID uuid.UUID) ([]ActiveRouteVehicle, error)
	FindOpenLap(ctx context.Context, vehicleID uuid.UUID) (*entities.Lap, error)
}

type routeLiveRepository struct {
	db *gorm.DB
}

func NewRouteLiveRepository(injector *do.Injector) (RouteLiveRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &routeLiveRepository{db: db}, nil
}

func (r *routeLiveRepository) FindActiveVehiclesForRoute(ctx context.Context, routeID uuid.UUID) ([]ActiveRouteVehicle, error) {
	var vehicles []ActiveRouteVehicle
	err := r.db.WithContext(ctx).
		Table("micros.vehicle_routes").
		Select("micros.vehicle_routes.vehicle_id AS vehicle_id, micros.vehicle_routes.pin_number AS pin_number, device_installations.imei AS imei").
		Joins("JOIN device_installations ON device_installations.vehicle_id = micros.vehicle_routes.vehicle_id AND device_installations.removed_at IS NULL AND device_installations.status = true").
		Where("micros.vehicle_routes.route_id = ? AND micros.vehicle_routes.active = ? AND micros.vehicle_routes.status = ?", routeID, true, true).
		Scan(&vehicles).Error
	return vehicles, err
}

func (r *routeLiveRepository) FindOpenLap(ctx context.Context, vehicleID uuid.UUID) (*entities.Lap, error) {
	var lap entities.Lap
	err := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND lap_status = ?", vehicleID, "IN_PROGRESS").
		Order("started_at DESC").
		First(&lap).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &lap, nil
}
