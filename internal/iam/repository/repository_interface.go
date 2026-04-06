package repository

import (
	"context"

	appcontext "MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/iam/models"
)

//go:generate mockgen -source=repository_interface.go -destination=../mocks/mock_repository.go -package=mocks

type IamStore interface {
	SaveSession(ctx *appcontext.AppContext, session *models.Session) error
}

type AudienceStore interface {
	IsValidAudience(ctx context.Context, audience string) (bool, error)
	GetAllowedAudiences(ctx context.Context) ([]string, error)
	GetUserAudiences(ctx context.Context, userUUID string) ([]string, error)
	RegisterAudience(ctx context.Context, audience string, metadata map[string]any) error
	RevokeAudience(ctx context.Context, audience string) error
	AssignAudienceToUser(ctx context.Context, userUUID, audience string) error
	UnassignAudienceFromUser(ctx context.Context, userUUID, audience string) error
}
