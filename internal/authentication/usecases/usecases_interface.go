package usecases

import (
	"context"

	"MydroX/anicetus/internal/authentication/dto"
)

//go:generate mockgen -source=usecases_interface.go -destination=../mocks/mock_usecases.go -package=mocks

type AuthenticationUsecases interface {
	Login(ctx context.Context, req *dto.LoginRequest) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error)
}
