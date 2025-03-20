package repository

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errors"
	"MydroX/anicetus/internal/users/models"
)

//go:generate mockgen -source=repository_interface.go -destination=../mocks/mock_repository.go -package=mocks

// UsersRepository is the interface to all the implemented db queries for every user related operation
type UsersRepository interface {
	// CreateUser is a method to create a user
	CreateUser(ctx *context.AppContext, user *models.User) *errors.Err

	// GetUserByUUID is a method to get a user by its uuid
	GetUserByUUID(ctx *context.AppContext, uuid string) (*models.User, *errors.Err)

	// UpdateUser is a method to update a user
	UpdateUser(ctx *context.AppContext, user *models.User) (*models.User, *errors.Err)

	// UpdatePassword is a method to update the password of a user
	UpdatePassword(ctx *context.AppContext, uuid, password string) *errors.Err

	// UpdateEmail is a method to update the email of a user
	UpdateEmail(ctx *context.AppContext, uuid, email string) *errors.Err

	// DeleteUser is a method to delete a user by its uuid
	DeleteUser(ctx *context.AppContext, uuid string) *errors.Err

	// GetUserByEmail is a method to get a user by its email
	GetUserByEmail(ctx *context.AppContext, email string) (*models.User, *errors.Err)

	// GetUserByUsername is a method to get a user by its username
	GetUserByUsername(ctx *context.AppContext, username string) (*models.User, *errors.Err)

	// GetUsers is a method to get all the users
	GetAllUsers(ctx *context.AppContext) ([]*models.User, *errors.Err)
}
