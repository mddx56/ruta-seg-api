package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"gorm.io/gorm"
)

type (
	UserRepository interface {
		Register(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error)
		GetUserById(ctx context.Context, tx *gorm.DB, userId string) (entities.User, error)
		GetUserByEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, error)
		GetUserByUsernameOrEmail(ctx context.Context, tx *gorm.DB, usernameOrEmail string) (entities.User, error)
		GetUserByGoogleID(ctx context.Context, tx *gorm.DB, googleID string) (entities.User, error)
		CheckEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, bool, error)
		CheckUsername(ctx context.Context, tx *gorm.DB, username string) (entities.User, bool, error)
		Update(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error)
		UpdatePassword(ctx context.Context, tx *gorm.DB, userId string, hashedPassword string) error
		UpdateBlockStatus(ctx context.Context, tx *gorm.DB, userId string, isBlocked bool) error
		UpdateStatus(ctx context.Context, tx *gorm.DB, userId string, status bool) error
		UpdateGoogleID(ctx context.Context, tx *gorm.DB, userID string, googleID string) error
	}

	userRepository struct {
		db *gorm.DB
	}
)

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Register(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) GetUserById(ctx context.Context, tx *gorm.DB, userId string) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).Where("id = ?", userId).Take(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).Where("email = ? AND status = ?", email, true).Take(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) GetUserByUsernameOrEmail(ctx context.Context, tx *gorm.DB, usernameOrEmail string) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).Where("(email = ? OR username = ?)", usernameOrEmail, usernameOrEmail).Take(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) CheckEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, bool, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).Where("email = ? AND status = ?", email, true).Take(&user).Error; err != nil {
		return entities.User{}, false, err
	}

	return user, true, nil
}

func (r *userRepository) CheckUsername(ctx context.Context, tx *gorm.DB, username string) (entities.User, bool, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).Where("username = ? AND status = ?", username, true).Take(&user).Error; err != nil {
		return entities.User{}, false, err
	}

	return user, true, nil
}

func (r *userRepository) Update(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Updates(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, tx *gorm.DB, userId string, hashedPassword string) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Model(&entities.User{}).Where("id = ? AND status = ?", userId, true).Update("password", hashedPassword).Error
}

func (r *userRepository) UpdateBlockStatus(ctx context.Context, tx *gorm.DB, userId string, isBlocked bool) error {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Model(&entities.User{}).Where("id = ?", userId).Update("is_blocked", isBlocked).Error; err != nil {
		return err
	}

	return nil
}

func (r *userRepository) UpdateStatus(ctx context.Context, tx *gorm.DB, userId string, status bool) error {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Model(&entities.User{}).Where("id = ?", userId).Update("status", status).Error; err != nil {
		return err
	}

	return nil
}

func (r *userRepository) GetUserByGoogleID(ctx context.Context, tx *gorm.DB, googleID string) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).Where("google_id = ?", googleID).Take(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) UpdateGoogleID(ctx context.Context, tx *gorm.DB, userID string, googleID string) error {
	if tx == nil {
		tx = r.db
	}

	return tx.WithContext(ctx).Model(&entities.User{}).Where("id = ?", userID).Update("google_id", googleID).Error
}
