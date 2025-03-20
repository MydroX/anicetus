package usecases

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errors"
	"MydroX/anicetus/internal/users/dto"
)

//go:generate mockgen -source=usecases_interface.go -destination=../mocks/mock_usecases.go -package=mocks

type UsersUsecases interface {
	Create(ctx *context.AppContext, user *dto.CreateUserRequest) *errors.Err
	Get(ctx *context.AppContext, uuid string) (*dto.GetUserResponse, *errors.Err)
	Update(ctx *context.AppContext, user *dto.UpdateUserRequest) *errors.Err
	UpdatePassword(ctx *context.AppContext, uuid string, password string) *errors.Err
	UpdateEmail(ctx *context.AppContext, uuid string, email string) *errors.Err
	Delete(ctx *context.AppContext, uuid string) *errors.Err
	GetAllUsers(ctx *context.AppContext) (*dto.GetAllUsersResponse, *errors.Err)
}
