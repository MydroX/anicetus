package usecases

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/users/dto"
)

//go:generate mockgen -source=usecases_interface.go -destination=../mocks/mock_usecases.go -package=mocks

type UsersUsecases interface {
	Create(ctx *context.AppContext, user *dto.CreateUserRequest) error
	Get(ctx *context.AppContext, uuid string) (*dto.GetUserResponse, error)
	Update(ctx *context.AppContext, user *dto.UpdateUserRequest) error
	UpdatePassword(ctx *context.AppContext, uuid string, password string) error
	UpdateEmail(ctx *context.AppContext, uuid string, email string) error
	Delete(ctx *context.AppContext, uuid string) error
	GetAllUsers(ctx *context.AppContext) (*dto.GetAllUsersResponse, error)
}
