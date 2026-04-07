package usecases

import (
	"context"

	"MydroX/anicetus/internal/services/dto"
	servicesrepository "MydroX/anicetus/internal/services/repository"
	"go.uber.org/zap"
)

type usecases struct {
	logger         *zap.SugaredLogger
	serviceStore   servicesrepository.ServiceStore
	serviceManager *ServiceManager
}

func New(l *zap.SugaredLogger, store servicesrepository.ServiceStore, manager *ServiceManager) ServicesUsecases {
	return &usecases{
		logger:         l,
		serviceStore:   store,
		serviceManager: manager,
	}
}

func (u *usecases) RegisterService(ctx context.Context, req *dto.RegisterServiceRequest) error {
	metadata := map[string]any{
		"service_name": req.ServiceName,
		"description":  req.Description,
		"permissions":  req.Permissions,
	}

	err := u.serviceStore.RegisterService(ctx, req.Audience, metadata)
	if err != nil {
		return err
	}

	u.serviceManager.InvalidateAllServicesCache(ctx)

	return nil
}

func (u *usecases) RevokeService(ctx context.Context, audience string) error {
	err := u.serviceStore.RevokeService(ctx, audience)
	if err != nil {
		return err
	}

	u.serviceManager.InvalidateAllServicesCache(ctx)

	return nil
}

func (u *usecases) GetAllServices(ctx context.Context) ([]string, error) {
	return u.serviceManager.GetAllowedServices(ctx)
}

func (u *usecases) GetUserServices(ctx context.Context, userUUID string) ([]string, error) {
	return u.serviceManager.GetUserServices(ctx, userUUID)
}

func (u *usecases) AssignServiceToUser(ctx context.Context, userUUID string, req *dto.AssignServiceRequest) error {
	err := u.serviceStore.AssignServiceToUser(ctx, userUUID, req.Audience)
	if err != nil {
		return err
	}

	u.serviceManager.InvalidateUserServicesCache(ctx, userUUID)

	return nil
}
