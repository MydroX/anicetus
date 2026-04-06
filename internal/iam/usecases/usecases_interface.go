package usecases

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/iam/dto"
)

//go:generate mockgen -source=usecases_interface.go -destination=../mocks/mock_usescases.go -package=mocks

type IamUsecasesService interface {
	Login(ctx *context.AppContext, req *dto.LoginRequest) (accessToken, refreshToken string, err error)
	Logout(ctx *context.AppContext, token string) error
	RefreshToken(ctx *context.AppContext, token string) (string, error)

	// Audience management
	RegisterAudience(ctx *context.AppContext, req *dto.RegisterAudienceRequest) error
	RevokeAudience(ctx *context.AppContext, audience string) error
	GetAllAudiences(ctx *context.AppContext) ([]string, error)
	GetUserAudiences(ctx *context.AppContext, userUUID string) ([]string, error)
	AssignAudienceToUser(ctx *context.AppContext, userUUID string, req *dto.AssignAudienceRequest) error
}
