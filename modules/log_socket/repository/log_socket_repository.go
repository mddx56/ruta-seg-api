package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"gorm.io/gorm"
)

type (
	LogSocketRepository interface {
		Create(ctx context.Context, tx *gorm.DB, logSocket entities.LogSocket) (entities.LogSocket, error)
	}

	logSocketRepository struct {
		db *gorm.DB
	}
)

func NewLogSocketRepository(db *gorm.DB) LogSocketRepository {
	return &logSocketRepository{
		db: db,
	}
}

func (r *logSocketRepository) Create(ctx context.Context, tx *gorm.DB, logSocket entities.LogSocket) (entities.LogSocket, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&logSocket).Error; err != nil {
		return entities.LogSocket{}, err
	}

	return logSocket, nil
}
