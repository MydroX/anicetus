package repository

import (
	"context"

	"MydroX/anicetus/internal/authentication/models"
)

//go:generate mockgen -source=repository_interface.go -destination=../mocks/mock_repository.go -package=mocks

type SessionStore interface {
	SaveSession(ctx context.Context, session *models.Session) error
	GetSessionByUUID(ctx context.Context, uuid string) (*models.Session, error)
	DeleteSession(ctx context.Context, uuid string) error
	DeleteAllUserSessions(ctx context.Context, userUUID string) error
	UpdateSessionLastUsed(ctx context.Context, uuid string) error
}
