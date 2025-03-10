package usecases

import (
	"MydroX/project-v/internal/users/dto"
	"context"
)

//go:generate mockgen -destination=../mocks/mock_usecases.go -package=mocks MydroX/project-v/internal/users/usecases UsersUsecases

type UsersUsecases interface {
	Create(ctx *context.Context, user *dto.CreateUserRequest) error
	Get(ctx *context.Context, uuid string) (*dto.GetUserResponse, error)
	Update(ctx *context.Context, user *dto.UpdateUserRequest) error
	UpdatePassword(ctx *context.Context, uuid string, password string) error
	UpdateEmail(ctx *context.Context, uuid string, email string) error
	Delete(ctx *context.Context, uuid string) error
	GetAllUsers(ctx *context.Context) (*dto.GetAllUsersResponse, error)
}
