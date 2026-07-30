package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type LapRepository interface {
	Create(ctx context.Context, lap *entities.Lap) error
	CloseLap(ctx context.Context, lap *entities.Lap) error
	FindOpenLap(ctx context.Context, vehicleID uuid.UUID) (*entities.Lap, error)
	FindActiveVehicleRoute(ctx context.Context, vehicleID uuid.UUID) (*entities.VehicleRoute, error)
	FindRouteByID(ctx context.Context, routeID uuid.UUID) (*entities.Route, error)
	FindGeofenceByID(ctx context.Context, id uuid.UUID) (*entities.Geofence, error)
	FindActiveFare(ctx context.Context, routeID uuid.UUID, at time.Time) (*entities.RouteFare, error)
	CreateCharge(ctx context.Context, charge *entities.LapCharge) error
	FindAll(ctx context.Context) ([]entities.Lap, error)
	FindByID(ctx context.Context, id uuid.UUID) (entities.Lap, error)
	FindVehicleOwnerID(ctx context.Context, vehicleID uuid.UUID) (uuid.UUID, error)
	UpdateTracking(ctx context.Context, lap *entities.Lap) error
	FindOpenLapsByRoute(ctx context.Context, routeID uuid.UUID, excludeVehicleID uuid.UUID) ([]entities.Lap, error)
}

type lapRepository struct {
	db *gorm.DB
}

func NewLapRepository(injector *do.Injector) (LapRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &lapRepository{db: db}, nil
}

func (r *lapRepository) Create(ctx context.Context, lap *entities.Lap) error {
	return r.db.WithContext(ctx).Create(lap).Error
}

func (r *lapRepository) CloseLap(ctx context.Context, lap *entities.Lap) error {
	return r.db.WithContext(ctx).
		Model(&entities.Lap{}).
		Where("id = ?", lap.ID).
		Select("ended_at", "duration_seconds", "lap_status", "updated_at").
		Updates(lap).Error
}

func (r *lapRepository) FindOpenLap(ctx context.Context, vehicleID uuid.UUID) (*entities.Lap, error) {
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

func (r *lapRepository) FindActiveVehicleRoute(ctx context.Context, vehicleID uuid.UUID) (*entities.VehicleRoute, error) {
	var vr entities.VehicleRoute
	err := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND active = ? AND status = ?", vehicleID, true, true).
		First(&vr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &vr, nil
}

func (r *lapRepository) FindRouteByID(ctx context.Context, routeID uuid.UUID) (*entities.Route, error) {
	var route entities.Route
	err := r.db.WithContext(ctx).Where("status = ?", true).First(&route, "id = ?", routeID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &route, nil
}

func (r *lapRepository) FindGeofenceByID(ctx context.Context, id uuid.UUID) (*entities.Geofence, error) {
	var geofence entities.Geofence
	err := r.db.WithContext(ctx).
		Preload("Points", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence ASC")
		}).
		Where("status = ?", true).
		First(&geofence, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &geofence, nil
}

func (r *lapRepository) FindActiveFare(ctx context.Context, routeID uuid.UUID, at time.Time) (*entities.RouteFare, error) {
	var fare entities.RouteFare
	err := r.db.WithContext(ctx).
		Where("route_id = ? AND status = ? AND effective_from <= ?", routeID, true, at).
		Order("effective_from DESC").
		First(&fare).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &fare, nil
}

func (r *lapRepository) CreateCharge(ctx context.Context, charge *entities.LapCharge) error {
	return r.db.WithContext(ctx).Create(charge).Error
}

func (r *lapRepository) FindAll(ctx context.Context) ([]entities.Lap, error) {
	var laps []entities.Lap
	err := r.db.WithContext(ctx).Preload("Charge").Order("started_at DESC").Find(&laps).Error
	return laps, err
}

func (r *lapRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Lap, error) {
	var lap entities.Lap
	err := r.db.WithContext(ctx).Preload("Charge").First(&lap, "id = ?", id).Error
	return lap, err
}

func (r *lapRepository) FindVehicleOwnerID(ctx context.Context, vehicleID uuid.UUID) (uuid.UUID, error) {
	var vehicle entities.Vehicle
	err := r.db.WithContext(ctx).Select("user_id").First(&vehicle, "id = ?", vehicleID).Error
	if err != nil {
		return uuid.Nil, err
	}
	return vehicle.UserID, nil
}

func (r *lapRepository) UpdateTracking(ctx context.Context, lap *entities.Lap) error {
	return r.db.WithContext(ctx).
		Model(&entities.Lap{}).
		Where("id = ?", lap.ID).
		Select(
			"last_movement_at",
			"last_stop_fine_at",
			"last_speed_fine_at",
			"last_polyline_index",
			"last_progress_meters",
			"last_progress_at",
			"overtaking_fined",
			"updated_at",
		).
		Updates(lap).Error
}

func (r *lapRepository) FindOpenLapsByRoute(ctx context.Context, routeID uuid.UUID, excludeVehicleID uuid.UUID) ([]entities.Lap, error) {
	var laps []entities.Lap
	err := r.db.WithContext(ctx).
		Where("route_id = ? AND lap_status = ? AND vehicle_id != ?", routeID, "IN_PROGRESS", excludeVehicleID).
		Find(&laps).Error
	return laps, err
}
