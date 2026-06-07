package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type GroupRepository interface {
	Create(ctx context.Context, group *entities.Group) error
	Update(ctx context.Context, group *entities.Group) error
	ChangeStatus(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (entities.Group, error)
	FindAll(ctx context.Context) ([]entities.Group, error)
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Group, error)
	AssignDevice(ctx context.Context, groupDevice *entities.GroupDevice) error
	RemoveDevice(ctx context.Context, groupID uuid.UUID, deviceIMEI string) error
	IsDeviceAssignedToGroup(ctx context.Context, groupID uuid.UUID, deviceIMEI string) (bool, error)
}

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(injector *do.Injector) (GroupRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &groupRepository{
		db: db,
	}, nil
}

func (r *groupRepository) Create(ctx context.Context, group *entities.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *groupRepository) Update(ctx context.Context, group *entities.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

func (r  *groupRepository) ChangeStatus(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Exec("UPDATE " + r.db.Statement.Table + " SET status = NOT status WHERE id = ?", id).Error
}

func (r *groupRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Group, error) {
	var group entities.Group
	err := r.db.WithContext(ctx).Where("id = ? AND status = ?", id, true).First(&group).Error
	return group, err
}

func (r *groupRepository) FindAll(ctx context.Context) ([]entities.Group, error) {
	var groups []entities.Group
	err := r.db.WithContext(ctx).Where("status = ?", true).Find(&groups).Error
	return groups, err
}

func (r *groupRepository) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Group, error) {
	var groups []entities.Group
	err := r.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, true).Find(&groups).Error
	return groups, err
}

func (r *groupRepository) AssignDevice(ctx context.Context, groupDevice *entities.GroupDevice) error {
	return r.db.WithContext(ctx).Create(groupDevice).Error
}

func (r *groupRepository) RemoveDevice(ctx context.Context, groupID uuid.UUID, deviceIMEI string) error {
	return r.db.WithContext(ctx).
		Model(&entities.GroupDevice{}).
		Where("group_id = ? AND device_imei = ?", groupID, deviceIMEI).
		Update("status", false).Error
}

func (r *groupRepository) IsDeviceAssignedToGroup(ctx context.Context, groupID uuid.UUID, deviceIMEI string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.GroupDevice{}).
		Where("group_id = ? AND device_imei = ? AND status = ?", groupID, deviceIMEI, true).
		Count(&count).Error
	return count > 0, err
}
