package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type MakeRepository interface {
	Create(ctx context.Context, make *entities.Make) error
	Update(ctx context.Context, make *entities.Make) error
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]entities.Make, error)
	FindByID(ctx context.Context, id uuid.UUID) (entities.Make, error)
}

type makeRepository struct {
	db *gorm.DB
}

func NewMakeRepository(injector *do.Injector) (MakeRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &makeRepository{
		db: db,
	}, nil
}

func (r *makeRepository) Create(ctx context.Context, make *entities.Make) error {
	return r.db.WithContext(ctx).Create(make).Error
}

func (r *makeRepository) Update(ctx context.Context, make *entities.Make) error {
	// Use Select to explicitly specify which fields to update
	// This prevents GORM from trying to update preloaded associations
	return r.db.WithContext(ctx).
		Model(make).
		Select("make_name", "updated_at").
		Updates(make).Error
}

func (r  *makeRepository) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Exec("UPDATE " + r.db.Statement.Table + " SET status = NOT status WHERE id = ?", id).Error
}

func (r *makeRepository) FindAll(ctx context.Context) ([]entities.Make, error) {
	var makes []entities.Make
	err := r.db.WithContext(ctx).Where("status = ?", true).Find(&makes).Error
	return makes, err
}

func (r *makeRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Make, error) {
	var make entities.Make
	err := r.db.WithContext(ctx).Where("status = ?", true).First(&make, id).Error
	return make, err
}
