package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type (
	AppVersionRepository interface {
		Create(ctx context.Context, tx *gorm.DB, appVersion entities.AppVersion) (entities.AppVersion, error)
		GetLatestVersion(ctx context.Context, tx *gorm.DB) (entities.AppVersion, error)
		GetAll(ctx context.Context, tx *gorm.DB) ([]entities.AppVersion, error)
		GetById(ctx context.Context, tx *gorm.DB, appId int) (entities.AppVersion, error)
		Update(ctx context.Context, tx *gorm.DB, appVersion entities.AppVersion) (entities.AppVersion, error)
		ChangeStatus(ctx context.Context, tx *gorm.DB, appId int) error
		Delete(ctx context.Context, tx *gorm.DB, appId int) error
	}

	appVersionRepository struct {
		db *gorm.DB
	}
)

func NewAppVersionRepository(injector *do.Injector) (AppVersionRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &appVersionRepository{
		db: db,
	}, nil
}

func (r *appVersionRepository) Create(ctx context.Context, tx *gorm.DB, appVersion entities.AppVersion) (entities.AppVersion, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&appVersion).Error; err != nil {
		return entities.AppVersion{}, err
	}

	return appVersion, nil
}

func (r *appVersionRepository) GetLatestVersion(ctx context.Context, tx *gorm.DB) (entities.AppVersion, error) {
	if tx == nil {
		tx = r.db
	}

	var appVersion entities.AppVersion
	err := tx.WithContext(ctx).
		Order("app_id DESC").
		First(&appVersion).Error

	if err != nil {
		return entities.AppVersion{}, err
	}

	return appVersion, nil
}

func (r *appVersionRepository) GetAll(ctx context.Context, tx *gorm.DB) ([]entities.AppVersion, error) {
	if tx == nil {
		tx = r.db
	}

	var appVersions []entities.AppVersion
	if err := tx.WithContext(ctx).Find(&appVersions).Error; err != nil {
		return nil, err
	}

	return appVersions, nil
}

func (r *appVersionRepository) GetById(ctx context.Context, tx *gorm.DB, appId int) (entities.AppVersion, error) {
	if tx == nil {
		tx = r.db
	}

	var appVersion entities.AppVersion
	if err := tx.WithContext(ctx).Where("app_id = ?", appId).First(&appVersion).Error; err != nil {
		return entities.AppVersion{}, err
	}

	return appVersion, nil
}

func (r *appVersionRepository) Update(ctx context.Context, tx *gorm.DB, appVersion entities.AppVersion) (entities.AppVersion, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Updates(&appVersion).Error; err != nil {
		return entities.AppVersion{}, err
	}

	return appVersion, nil
}

func (r *appVersionRepository) Delete(ctx context.Context, tx *gorm.DB, appId int) error {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Delete(&entities.AppVersion{}, appId).Error; err != nil {
		return err
	}

	return nil
}

func (r *appVersionRepository) ChangeStatus(ctx context.Context, tx *gorm.DB, appId int) error {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Exec("UPDATE app_versions SET status = NOT status WHERE app_id = ?", appId).Error; err != nil {
		return err
	}

	return nil
}
