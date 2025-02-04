package repository

import (
	"MydroX/project-v/internal/gateway/users/models"
	"context"
)

//go:generate mockgen -destination=../mocks/mock_repository.go -imports=models=MydroX/project-v/internal/users/models -package=mocks MydroX/project-v/internal/gateway/users/repository UsersRepository

// UsersRepository is the interface to all the implemented db queries for every user related operation
type UsersRepository interface {
	// CreateUser is a method to create a user
	CreateUser(ctx *context.Context, user *models.User) error

	// GetUserByUUID is a method to get a user by its uuid
	GetUserByUUID(ctx *context.Context, uuid string) (*models.User, error)

	// UpdateUser is a method to update a user
	UpdateUser(ctx *context.Context, user *models.User) (*models.User, error)

	// UpdatePassword is a method to update the password of a user
	UpdatePassword(ctx *context.Context, uuid, password string) error

	// UpdateEmail is a method to update the email of a user
	UpdateEmail(ctx *context.Context, uuid, email string) error

	// DeleteUser is a method to delete a user by its uuid
	DeleteUser(ctx *context.Context, uuid string) error

	// GetUserByEmail is a method to get a user by its email
	GetUserByEmail(ctx *context.Context, email string) (*models.User, error)

	// GetUserByUsername is a method to get a user by its username
	GetUserByUsername(ctx *context.Context, username string) (*models.User, error)
}
