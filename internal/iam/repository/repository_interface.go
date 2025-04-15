package repository

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/iam/models"
)

//go:generate mockgen -source=repository_interface.go -destination=../mocks/mock_repository.go -package=mocks

type IamStore interface {
	SaveSession(ctx *context.AppContext, session *models.Session) error
}

type AudienceStore interface {
	IsValidAudience(audience string) (bool, error)
	GetAllowedAudiences() ([]string, error)
	RegisterAudience(audience string, metadata map[string]any) error
	RevokeAudience(audience string) error
}
