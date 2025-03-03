package usecases

import (
	"MydroX/project-v/internal/users/dto"
	"context"
)

//go:generate mockgen -destination=../mocks/mock_usecases.go -package=mocks MydroX/project-v/internal/users/usecases UsersUsecases

// UsersUsecases is the interface to all the implemented usecases for the users entity
type UsersUsecases interface {
	// Create is the usecase to create a new user
	Create(ctx *context.Context, user *dto.CreateUserRequest) error

	// Get is the usecase to get a user by its uuid
	Get(ctx *context.Context, uuid string) (*dto.GetUserResponse, error)

	// Update is the usecase to update a user
	Update(ctx *context.Context, user *dto.UpdateUserRequest) error

	// UpdatePassword is the usecase to update the password of a user
	UpdatePassword(ctx *context.Context, uuid string, password string) error

	// UpdateEmail is the usecase to update the email of a user
	UpdateEmail(ctx *context.Context, uuid string, email string) error

	// Delete is the usecase to delete a user
	Delete(ctx *context.Context, uuid string) error

	// Login is the usecase to login a user
	Login(ctx *context.Context, username, email, password string) (string, error)

	// GetUsers is the usecase to get all the users
	GetAllUsers(ctx *context.Context) (*dto.GetAllUsersResponse, error)
}
