package usecases

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errors"
	"MydroX/anicetus/internal/iam/dto"
)

//go:generate mockgen -source=usecases_interface.go -destination=../mocks/mock_usescases.go -package=mocks

type IamUsecasesInterface interface {
	Login(ctx *context.AppContext, req *dto.LoginRequest) (accessToken, refreshToken string, err *errors.Err)
	Logout(ctx *context.AppContext, token string) *errors.Err
	RefreshToken(ctx *context.AppContext, token string) (string, *errors.Err)
}
