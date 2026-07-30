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

type FineRepository interface {
	Create(ctx context.Context, fine *entities.Fine) error
	UpdateStatus(ctx context.Context, id uuid.UUID, fineStatus string, notes *string) error
	FindAll(ctx context.Context) ([]entities.Fine, error)
	FindAllByOwner(ctx context.Context, ownerUserID uuid.UUID) ([]entities.Fine, error)
	FindByID(ctx context.Context, id uuid.UUID) (entities.Fine, error)
	FindVehicleOwnerID(ctx context.Context, vehicleID uuid.UUID) (uuid.UUID, error)
}

type FineTypeRepository interface {
	FindAll(ctx context.Context) ([]entities.FineType, error)
	FindByCode(ctx context.Context, code string) (entities.FineType, error)
}

type fineRepository struct {
	db *gorm.DB
}

func NewFineRepository(injector *do.Injector) (FineRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &fineRepository{db: db}, nil
}

func (r *fineRepository) Create(ctx context.Context, fine *entities.Fine) error {
	return r.db.WithContext(ctx).Create(fine).Error
}

func (r *fineRepository) UpdateStatus(ctx context.Context, id uuid.UUID, fineStatus string, notes *string) error {
	updates := map[string]interface{}{"fine_status": fineStatus}
	if notes != nil {
		updates["notes"] = *notes
	}
	return r.db.WithContext(ctx).
		Model(&entities.Fine{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *fineRepository) FindAll(ctx context.Context) ([]entities.Fine, error) {
	var fines []entities.Fine
	err := r.db.WithContext(ctx).
		Preload("FineType").
		Where("status = ?", true).
		Order("occurred_at DESC").
		Find(&fines).Error
	return fines, err
}

func (r *fineRepository) FindAllByOwner(ctx context.Context, ownerUserID uuid.UUID) ([]entities.Fine, error) {
	var fines []entities.Fine
	err := r.db.WithContext(ctx).
		Preload("FineType").
		Joins("JOIN vehicles ON vehicles.id = fines.vehicle_id").
		Where("fines.status = ? AND vehicles.user_id = ?", true, ownerUserID).
		Order("fines.occurred_at DESC").
		Find(&fines).Error
	return fines, err
}

func (r *fineRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Fine, error) {
	var fine entities.Fine
	err := r.db.WithContext(ctx).
		Preload("FineType").
		Where("status = ?", true).
		First(&fine, "id = ?", id).Error
	return fine, err
}

func (r *fineRepository) FindVehicleOwnerID(ctx context.Context, vehicleID uuid.UUID) (uuid.UUID, error) {
	var vehicle entities.Vehicle
	err := r.db.WithContext(ctx).Select("user_id").First(&vehicle, "id = ?", vehicleID).Error
	if err != nil {
		return uuid.Nil, err
	}
	return vehicle.UserID, nil
}

type fineTypeRepository struct {
	db *gorm.DB
}

func NewFineTypeRepository(injector *do.Injector) (FineTypeRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &fineTypeRepository{db: db}, nil
}

func (r *fineTypeRepository) FindAll(ctx context.Context) ([]entities.FineType, error) {
	var fineTypes []entities.FineType
	err := r.db.WithContext(ctx).Where("status = ?", true).Order("code ASC").Find(&fineTypes).Error
	return fineTypes, err
}

func (r *fineTypeRepository) FindByCode(ctx context.Context, code string) (entities.FineType, error) {
	var fineType entities.FineType
	err := r.db.WithContext(ctx).Where("status = ? AND code = ?", true, code).First(&fineType).Error
	if err != nil {
		return entities.FineType{}, errors.New("tipo de multa no encontrado: " + code)
	}
	return fineType, nil
}
