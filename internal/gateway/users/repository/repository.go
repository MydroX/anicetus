package repository

import (
	"MydroX/project-v/internal/common/errors"
	"MydroX/project-v/internal/gateway/users/models"
	"MydroX/project-v/pkg/logger"
	"context"
	"fmt"

	"gorm.io/gorm"
)

type repository struct {
	logger *logger.Logger
	db     *gorm.DB
}

// NewRepository is creating an interface for every method of the repository
func NewRepository(l *logger.Logger, db *gorm.DB) UsersRepository {
	return &repository{
		logger: l,
		db:     db,
	}
}

func (r *repository) CreateUser(ctx *context.Context, user *models.User) error {
	res := r.db.Create(&user)
	if res.Error != nil {
		if res.Error == gorm.ErrDuplicatedKey {
			*ctx = context.WithValue(*ctx, errors.CtxErrorCodeKey, errors.CODE_DUPLICATE_ENTITY)
			return fmt.Errorf("user already exists")
		}
		r.logger.Zap.Sugar().Errorf("error creating user: %v", res.Error)
		return res.Error
	}

	return nil
}

func (r *repository) GetUserByUUID(ctx *context.Context, uuid string) (*models.User, error) {
	var user models.User

	res := r.db.First(&user, uuid)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			*ctx = context.WithValue(*ctx, errors.CtxErrorCodeKey, errors.CODE_ENTITY_NOT_FOUND)
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Zap.Sugar().Errorf("error getting user: %v", res.Error)
		return nil, res.Error
	}

	return &user, nil
}

func (r *repository) UpdateUser(_ *context.Context, user *models.User) error {
	res := r.db.Save(&user)
	if res.Error != nil {
		r.logger.Zap.Sugar().Errorf("error updating user: %v", res.Error)
		return res.Error
	}

	return nil
}

func (r *repository) UpdatePassword(_ *context.Context, uuid, password string) error {
	user := models.User{
		UUID: uuid,
	}
	res := r.db.Model(&user).Update("password", password)
	if res.Error != nil {
		r.logger.Zap.Sugar().Errorf("error updating user password: %v", res.Error)
		return res.Error
	}
	return nil
}

func (r *repository) UpdateEmail(_ *context.Context, uuid, email string) error {
	user := models.User{
		UUID: uuid,
	}
	res := r.db.Model(&user).Update("email", email)
	if res.Error != nil {
		r.logger.Zap.Sugar().Errorf("error updating user email: %v", res.Error)
		return res.Error
	}
	return nil
}

func (r *repository) DeleteUser(_ *context.Context, uuid string) error {
	res := r.db.Delete(&models.User{}, uuid)
	if res.Error != nil {
		r.logger.Zap.Sugar().Errorf("error deleting user: %v", res.Error)
		return res.Error
	}

	return nil
}

func (r *repository) GetUserByEmail(ctx *context.Context, email string) (*models.User, error) {
	user := models.User{
		Email: email,
	}

	res := r.db.Where("email = ?", email).First(&user)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			*ctx = context.WithValue(*ctx, errors.CtxErrorCodeKey, errors.CODE_ENTITY_NOT_FOUND)
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Zap.Sugar().Errorf("error getting user by email: %v", res.Error)
		return nil, res.Error
	}

	return &user, nil
}

func (r *repository) GetUserByUsername(ctx *context.Context, username string) (*models.User, error) {
	user := models.User{
		Username: username,
	}

	res := r.db.Where("username = ?", username).First(&user)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			*ctx = context.WithValue(*ctx, errors.CtxErrorCodeKey, errors.CODE_ENTITY_NOT_FOUND)
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Zap.Sugar().Errorf("error getting user by username: %v", res.Error)
		return nil, res.Error
	}

	return &user, nil
}
