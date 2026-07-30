package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type GeofenceRepository interface {
	Create(ctx context.Context, geofence *entities.Geofence) error
	Update(ctx context.Context, geofence *entities.Geofence, points []entities.GeofencePoint) error
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]entities.Geofence, error)
	FindByID(ctx context.Context, id uuid.UUID) (entities.Geofence, error)
}

type geofenceRepository struct {
	db *gorm.DB
}

func NewGeofenceRepository(injector *do.Injector) (GeofenceRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &geofenceRepository{
		db: db,
	}, nil
}

func (r *geofenceRepository) Create(ctx context.Context, geofence *entities.Geofence) error {
	return r.db.WithContext(ctx).Create(geofence).Error
}

// Update actualiza los campos base de la geocerca y, si points no es nil, reemplaza
// la lista completa de puntos (borra los existentes e inserta los nuevos).
func (r *geofenceRepository) Update(ctx context.Context, geofence *entities.Geofence, points []entities.GeofencePoint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entities.Geofence{}).
			Where("id = ?", geofence.ID).
			Select("name", "type", "radius", "updated_at").
			Updates(geofence).Error; err != nil {
			return err
		}

		if points == nil {
			return nil
		}

		if err := tx.Where("geofence_id = ?", geofence.ID).Delete(&entities.GeofencePoint{}).Error; err != nil {
			return err
		}

		if len(points) > 0 {
			if err := tx.Create(&points).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *geofenceRepository) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entities.Geofence{}).
		Where("id = ?", id).
		Update("status", gorm.Expr("NOT status")).Error
}

func (r *geofenceRepository) FindAll(ctx context.Context) ([]entities.Geofence, error) {
	var geofences []entities.Geofence
	err := r.db.WithContext(ctx).
		Preload("Points", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence ASC")
		}).
		Where("status = ?", true).
		Order("name ASC").
		Find(&geofences).Error
	return geofences, err
}

func (r *geofenceRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Geofence, error) {
	var geofence entities.Geofence
	err := r.db.WithContext(ctx).
		Preload("Points", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence ASC")
		}).
		Where("status = ?", true).
		First(&geofence, "id = ?", id).Error
	return geofence, err
}
