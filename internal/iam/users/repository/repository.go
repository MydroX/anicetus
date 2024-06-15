// Package repository provides every implementation of database operations.
package repository

import (
	"MydroX/project-v/internal/iam/users/models"
	apiError "MydroX/project-v/pkg/errors"
	"MydroX/project-v/pkg/logger"
	"fmt"

	"github.com/google/uuid"
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

func (r *repository) CreateUser(user *models.User) error {
	res := r.db.Create(&user)
	if res.Error != nil {
		if res.Error == gorm.ErrDuplicatedKey {
			fmt.Println(res.Error)
		}
		r.logger.Zap.Sugar().Errorf("error creating user: %v", res.Error)
		return res.Error
	}

	return nil
}

func (r *repository) GetUser(uuid uuid.UUID) (*models.User, error) {
	var user models.User

	res := r.db.First(&user, uuid)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, apiError.ErrNotFound
		}
		r.logger.Zap.Sugar().Errorf("error getting user: %v", res.Error)
		return nil, res.Error
	}

	return &user, nil
}

func (r *repository) UpdateUser(user *models.User) error {
	res := r.db.Save(&user)
	if res.Error != nil {
		r.logger.Zap.Sugar().Errorf("error updating user: %v", res.Error)
		return res.Error
	}

	return nil
}

func (r *repository) UpdatePassword(uuid uuid.UUID, password string) error {
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

func (r *repository) UpdateEmail(uuid uuid.UUID, email string) error {
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

func (r *repository) DeleteUser(uuid uuid.UUID) error {
	res := r.db.Delete(&models.User{}, uuid)
	if res.Error != nil {
		r.logger.Zap.Sugar().Errorf("error deleting user: %v", res.Error)
		return res.Error
	}

	return nil
}
