package repository

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/users/models"
)

//go:generate mockgen -source=repository_interface.go -destination=../mocks/mock_repository.go -package=mocks

// UsersRepository is the interface to all the implemented db queries for every user related operation
type UsersRepository interface {
	// CreateUser is a method to create a user
	CreateUser(ctx *context.AppContext, user *models.User) error

	// GetUserByUUID is a method to get a user by its uuid
	GetUserByUUID(ctx *context.AppContext, uuid string) (*models.User, error)

	// UpdateUser is a method to update a user
	UpdateUser(ctx *context.AppContext, user *models.User) (*models.User, error)

	// UpdatePassword is a method to update the password of a user
	UpdatePassword(ctx *context.AppContext, uuid, password string) error

	// UpdateEmail is a method to update the email of a user
	UpdateEmail(ctx *context.AppContext, uuid, email string) error

	// DeleteUser is a method to delete a user by its uuid
	DeleteUser(ctx *context.AppContext, uuid string) error

	// GetUserByEmail is a method to get a user by its email
	GetUserByEmail(ctx *context.AppContext, email string) (*models.User, error)

	// GetUserByUsername is a method to get a user by its username
	GetUserByUsername(ctx *context.AppContext, username string) (*models.User, error)

	// GetUsers is a method to get all the users
	GetAllUsers(ctx *context.AppContext) ([]*models.User, error)
}
