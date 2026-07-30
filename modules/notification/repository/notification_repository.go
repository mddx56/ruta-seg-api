package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DeviceTokenRepository interface {
	Upsert(ctx context.Context, token *entities.UserDeviceToken) error
	FindTokensByUserID(ctx context.Context, userID uuid.UUID) ([]string, error)
	DeleteByToken(ctx context.Context, token string) error
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *entities.Notification) error
	FindAllByUser(ctx context.Context, userID uuid.UUID) ([]entities.Notification, error)
	MarkRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type deviceTokenRepository struct {
	db *gorm.DB
}

func NewDeviceTokenRepository(injector *do.Injector) (DeviceTokenRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &deviceTokenRepository{db: db}, nil
}

// Upsert registra el token o, si ya existía (p.ej. el dueño reinstaló la app y a otro
// usuario le habían asignado antes ese mismo token), lo reasigna al usuario actual.
func (r *deviceTokenRepository) Upsert(ctx context.Context, token *entities.UserDeviceToken) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "token"}},
			DoUpdates: clause.AssignmentColumns([]string{"user_id", "platform", "updated_at", "status"}),
		}).
		Create(token).Error
}

func (r *deviceTokenRepository) FindTokensByUserID(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var tokens []string
	err := r.db.WithContext(ctx).
		Model(&entities.UserDeviceToken{}).
		Where("user_id = ? AND status = ?", userID, true).
		Pluck("token", &tokens).Error
	return tokens, err
}

func (r *deviceTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&entities.UserDeviceToken{}).Error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(injector *do.Injector) (NotificationRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &notificationRepository{db: db}, nil
}

func (r *notificationRepository) Create(ctx context.Context, notification *entities.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *notificationRepository) FindAllByUser(ctx context.Context, userID uuid.UUID) ([]entities.Notification, error) {
	var notifications []entities.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, true).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) MarkRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entities.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("read", true).Error
}
