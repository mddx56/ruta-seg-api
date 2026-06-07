package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type ModelRepository interface {
	Create(ctx context.Context, model *entities.Model) error
	Update(ctx context.Context, model *entities.Model) error
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]entities.Model, error)
	FindByID(ctx context.Context, id uuid.UUID) (entities.Model, error)
}

type modelRepository struct {
	db *gorm.DB
}

func NewModelRepository(injector *do.Injector) (ModelRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &modelRepository{
		db: db,
	}, nil
}

func (r *modelRepository) Create(ctx context.Context, model *entities.Model) error {
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *modelRepository) Update(ctx context.Context, model *entities.Model) error {
	// Use Select to explicitly specify which fields to update
	// This prevents GORM from trying to update preloaded associations
	return r.db.WithContext(ctx).
		Model(model).
		Select("model_name", "vehicle_type_id", "make_id", "updated_at").
		Updates(model).Error
}

func (r  *modelRepository) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Exec("UPDATE " + r.db.Statement.Table + " SET status = NOT status WHERE id = ?", id).Error
}

func (r *modelRepository) FindAll(ctx context.Context) ([]entities.Model, error) {
	var models []entities.Model
	err := r.db.WithContext(ctx).
		Preload("VehicleType").
		Preload("Make").
		Where("status = ?", true).
		Find(&models).Error
	return models, err
}

func (r *modelRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Model, error) {
	var model entities.Model
	err := r.db.WithContext(ctx).
		Preload("VehicleType").
		Preload("Make").
		Where("status = ?", true).
		First(&model, id).Error
	return model, err
}
