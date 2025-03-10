package usecases

import "context"

//go:generate mockgen -destination=../mocks/mock_usecases.go -package=mocks MydroX/project-v/internal/iam/usecases IamUsecasesInterface

type IamUsecasesInterface interface {
	Login(ctx *context.Context, username, email, password string) (accessToken, refreshToken string, err error)
	Logout(ctx *context.Context, token string) error
	RefreshToken(ctx *context.Context, token string) (string, error)
}
