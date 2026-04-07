package usecases

import (
	"context"

	"MydroX/anicetus/internal/services/dto"
)

//go:generate mockgen -source=usecases_interface.go -destination=../mocks/mock_usecases.go -package=mocks

type ServicesUsecases interface {
	RegisterService(ctx context.Context, req *dto.RegisterServiceRequest) error
	RevokeService(ctx context.Context, audience string) error
	GetAllServices(ctx context.Context) ([]string, error)
	GetUserServices(ctx context.Context, userUUID string) ([]string, error)
	AssignServiceToUser(ctx context.Context, userUUID string, req *dto.AssignServiceRequest) error
}
