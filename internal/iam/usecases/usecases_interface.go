package usecases

import (
	"context"

	"MydroX/anicetus/internal/iam/dto"
)

//go:generate mockgen -source=usecases_interface.go -destination=../mocks/mock_usescases.go -package=mocks

type IamUsecasesService interface {
	Login(ctx context.Context, req *dto.LoginRequest) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context, token string) error
	RefreshToken(ctx context.Context, token string) (string, error)

	// Audience management
	RegisterAudience(ctx context.Context, req *dto.RegisterAudienceRequest) error
	RevokeAudience(ctx context.Context, audience string) error
	GetAllAudiences(ctx context.Context) ([]string, error)
	GetUserAudiences(ctx context.Context, userUUID string) ([]string, error)
	AssignAudienceToUser(ctx context.Context, userUUID string, req *dto.AssignAudienceRequest) error
}
