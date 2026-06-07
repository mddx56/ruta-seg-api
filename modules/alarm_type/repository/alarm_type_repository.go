package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"gorm.io/gorm"
)

type AlarmTypeRepository interface {
	Create(ctx context.Context, alarmType *entities.AlarmType) error
	FindAll(ctx context.Context) ([]entities.AlarmType, error)
}

type alarmTypeRepository struct {
	db *gorm.DB
}

func NewAlarmTypeRepository(db *gorm.DB) AlarmTypeRepository {
	return &alarmTypeRepository{
		db: db,
	}
}

func (r *alarmTypeRepository) Create(ctx context.Context, alarmType *entities.AlarmType) error {
	return r.db.WithContext(ctx).Create(alarmType).Error
}

func (r *alarmTypeRepository) FindAll(ctx context.Context) ([]entities.AlarmType, error) {
	var alarmTypes []entities.AlarmType
	err := r.db.WithContext(ctx).Find(&alarmTypes).Error
	return alarmTypes, err
}
