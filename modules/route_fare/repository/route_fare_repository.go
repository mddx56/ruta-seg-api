package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type RouteFareRepository interface {
	Create(ctx context.Context, fare *entities.RouteFare) error
	Update(ctx context.Context, fare *entities.RouteFare) error
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]entities.RouteFare, error)
	FindByID(ctx context.Context, id uuid.UUID) (entities.RouteFare, error)
}

type routeFareRepository struct {
	db *gorm.DB
}

func NewRouteFareRepository(injector *do.Injector) (RouteFareRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &routeFareRepository{
		db: db,
	}, nil
}

func (r *routeFareRepository) Create(ctx context.Context, fare *entities.RouteFare) error {
	return r.db.WithContext(ctx).Create(fare).Error
}

func (r *routeFareRepository) Update(ctx context.Context, fare *entities.RouteFare) error {
	return r.db.WithContext(ctx).
		Model(&entities.RouteFare{}).
		Where("id = ?", fare.ID).
		Select("amount_per_lap", "effective_from", "updated_at").
		Updates(fare).Error
}

func (r *routeFareRepository) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entities.RouteFare{}).
		Where("id = ?", id).
		Update("status", gorm.Expr("NOT status")).Error
}

func (r *routeFareRepository) FindAll(ctx context.Context) ([]entities.RouteFare, error) {
	var fares []entities.RouteFare
	err := r.db.WithContext(ctx).
		Where("status = ?", true).
		Order("effective_from DESC").
		Find(&fares).Error
	return fares, err
}

func (r *routeFareRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.RouteFare, error) {
	var fare entities.RouteFare
	err := r.db.WithContext(ctx).
		Where("status = ?", true).
		First(&fare, "id = ?", id).Error
	return fare, err
}
