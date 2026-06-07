package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"gorm.io/gorm"
)

type AlarmRuleRepository interface {
	Create(ctx context.Context, alarmRule *entities.AlarmRule) error
	FindAll(ctx context.Context) ([]entities.AlarmRule, error)
}

type alarmRuleRepository struct {
	db *gorm.DB
}

func NewAlarmRuleRepository(db *gorm.DB) AlarmRuleRepository {
	return &alarmRuleRepository{
		db: db,
	}
}

func (r *alarmRuleRepository) Create(ctx context.Context, alarmRule *entities.AlarmRule) error {
	return r.db.WithContext(ctx).Create(alarmRule).Error
}

func (r *alarmRuleRepository) FindAll(ctx context.Context) ([]entities.AlarmRule, error) {
	var alarmRules []entities.AlarmRule
	err := r.db.WithContext(ctx).Preload("AlarmType").Preload("Devices").Find(&alarmRules).Error
	return alarmRules, err
}
