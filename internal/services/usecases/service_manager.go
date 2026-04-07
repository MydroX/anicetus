package usecases

import (
	"context"
	"encoding/json"
	"fmt"

	servicesrepository "MydroX/anicetus/internal/services/repository"
	"github.com/valkey-io/valkey-go"
	"go.uber.org/zap"
)

const (
	serviceCacheKey     = "allowed_services"
	userServiceCacheKey = "user_services"
)

// ServiceManager handles caching and retrieval of allowed JWT audiences (services)
type ServiceManager struct {
	logger     *zap.SugaredLogger
	repository servicesrepository.ServiceStore
	valkey     valkey.Client
}

// NewServiceManager creates a new service manager
func NewServiceManager(logger *zap.SugaredLogger, repo servicesrepository.ServiceStore, client valkey.Client) *ServiceManager {
	return &ServiceManager{
		logger:     logger,
		repository: repo,
		valkey:     client,
	}
}

// CacheAllowedServices loads services from database into cache
func (sm *ServiceManager) CacheAllowedServices(ctx context.Context) error {
	services, err := sm.repository.GetAllowedServices(ctx)
	if err != nil {
		return err
	}

	return sm.setSliceCache(ctx, serviceCacheKey, services)
}

// GetAllowedServices retrieves services from cache or database
func (sm *ServiceManager) GetAllowedServices(ctx context.Context) ([]string, error) {
	if services, ok := sm.getSliceCache(ctx, serviceCacheKey); ok {
		return services, nil
	}

	services, err := sm.repository.GetAllowedServices(ctx)
	if err != nil {
		return nil, err
	}

	go func() { //nolint:gosec // Background context is correct here - goroutine outlives the request
		if cacheErr := sm.CacheAllowedServices(context.Background()); cacheErr != nil {
			sm.logger.Warnw("Failed to update services cache", "error", cacheErr)
		}
	}()

	return services, nil
}

// GetUserServices retrieves services for a specific user from cache or database
func (sm *ServiceManager) GetUserServices(ctx context.Context, userUUID string) ([]string, error) {
	cacheKey := fmt.Sprintf("%s:%s", userServiceCacheKey, userUUID)

	if services, ok := sm.getSliceCache(ctx, cacheKey); ok {
		return services, nil
	}

	services, err := sm.repository.GetUserServices(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	_ = sm.setSliceCache(ctx, cacheKey, services)

	return services, nil
}

// InvalidateUserServicesCache removes a user's service cache entry
func (sm *ServiceManager) InvalidateUserServicesCache(ctx context.Context, userUUID string) {
	cacheKey := fmt.Sprintf("%s:%s", userServiceCacheKey, userUUID)

	sm.valkey.Do(ctx, sm.valkey.B().Del().Key(cacheKey).Build())
}

// InvalidateAllServicesCache removes the global services cache entry
func (sm *ServiceManager) InvalidateAllServicesCache(ctx context.Context) {
	sm.valkey.Do(ctx, sm.valkey.B().Del().Key(serviceCacheKey).Build())
}

func (sm *ServiceManager) setSliceCache(ctx context.Context, key string, values []string) error {
	data, err := json.Marshal(values)
	if err != nil {
		return err
	}

	return sm.valkey.Do(ctx, sm.valkey.B().Set().Key(key).Value(string(data)).Build()).Error()
}

func (sm *ServiceManager) getSliceCache(ctx context.Context, key string) ([]string, bool) {
	result, err := sm.valkey.Do(ctx, sm.valkey.B().Get().Key(key).Build()).ToString()
	if err != nil {
		return nil, false
	}

	var services []string
	if err := json.Unmarshal([]byte(result), &services); err != nil {
		sm.logger.Warnw("Failed to unmarshal cache value", "key", key, "error", err)

		return nil, false
	}

	return services, true
}
