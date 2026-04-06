package usecases

import (
	"context"
	"MydroX/anicetus/internal/users/dto"
)

//go:generate mockgen -source=usecases_interface.go -destination=../mocks/mock_usecases.go -package=mocks

type UsersUsecases interface {
	Create(ctx context.Context, user *dto.CreateUserRequest) error
	Get(ctx context.Context, uuid string) (*dto.GetUserResponse, error)
	Update(ctx context.Context, user *dto.UpdateUserRequest) error
	UpdatePassword(ctx context.Context, uuid string, password string) error
	UpdateEmail(ctx context.Context, uuid string, email string) error
	Delete(ctx context.Context, uuid string) error
	GetAllUsers(ctx context.Context) (*dto.GetAllUsersResponse, error)
}
