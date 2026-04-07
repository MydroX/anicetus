package repository

import "context"

//go:generate mockgen -source=repository_interface.go -destination=../mocks/mock_repository.go -package=mocks

type ServiceStore interface {
	IsValidService(ctx context.Context, audience string) (bool, error)
	GetAllowedServices(ctx context.Context) ([]string, error)
	GetUserServices(ctx context.Context, userUUID string) ([]string, error)
	RegisterService(ctx context.Context, audience string, metadata map[string]any) error
	RevokeService(ctx context.Context, audience string) error
	AssignServiceToUser(ctx context.Context, userUUID, audience string) error
	UnassignServiceFromUser(ctx context.Context, userUUID, audience string) error
}
