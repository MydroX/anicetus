package repository

import (
	"context"

	"MydroX/anicetus/internal/users/models"
)

//go:generate mockgen -source=repository_interface.go -destination=../mocks/mock_repository.go -package=mocks

type UsersRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByUUID(ctx context.Context, uuid string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	UpdatePassword(ctx context.Context, uuid, password string) error
	UpdateEmail(ctx context.Context, uuid, email string) error
	DeleteUser(ctx context.Context, uuid string) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]*models.User, error)
}
