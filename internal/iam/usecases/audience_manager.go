package usecases

import (
	"context"
	"encoding/json"
	"fmt"

	iamrepository "MydroX/anicetus/internal/iam/repository"

	"github.com/dgraph-io/ristretto/v2"
	"go.uber.org/zap"
)

const (
	audienceCacheKey     = "allowed_audiences"
	userAudienceCacheKey = "user_audiences"
)

// AudienceManager handles caching and retrieval of allowed JWT audiences
type AudienceManager struct {
	logger      *zap.SugaredLogger
	repository  iamrepository.AudienceStore
	cacheClient *ristretto.Cache[string, string]
}

// NewAudienceManager creates a new audience manager
func NewAudienceManager(logger *zap.SugaredLogger, repo iamrepository.AudienceStore, cache *ristretto.Cache[string, string]) *AudienceManager {
	return &AudienceManager{
		logger:      logger,
		repository:  repo,
		cacheClient: cache,
	}
}

// CacheAllowedAudiences loads audiences from database into cache
func (am *AudienceManager) CacheAllowedAudiences(ctx context.Context) error {
	audiences, err := am.repository.GetAllowedAudiences(ctx)
	if err != nil {
		return err
	}

	return am.setSliceCache(audienceCacheKey, audiences)
}

// GetAllowedAudiences retrieves audiences from cache or database
func (am *AudienceManager) GetAllowedAudiences(ctx context.Context) ([]string, error) {
	if audiences, ok := am.getSliceCache(audienceCacheKey); ok {
		return audiences, nil
	}

	audiences, err := am.repository.GetAllowedAudiences(ctx)
	if err != nil {
		return nil, err
	}

	go func() { //nolint:gosec // Background context is correct here - goroutine outlives the request
		if cacheErr := am.CacheAllowedAudiences(context.Background()); cacheErr != nil {
			am.logger.Warnw("Failed to update audiences cache", "error", cacheErr)
		}
	}()

	return audiences, nil
}

// GetUserAudiences retrieves audiences for a specific user from cache or database
func (am *AudienceManager) GetUserAudiences(ctx context.Context, userUUID string) ([]string, error) {
	cacheKey := fmt.Sprintf("%s:%s", userAudienceCacheKey, userUUID)

	if audiences, ok := am.getSliceCache(cacheKey); ok {
		return audiences, nil
	}

	audiences, err := am.repository.GetUserAudiences(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	_ = am.setSliceCache(cacheKey, audiences)

	return audiences, nil
}

// InvalidateUserAudiencesCache removes a user's audience cache entry
func (am *AudienceManager) InvalidateUserAudiencesCache(userUUID string) {
	cacheKey := fmt.Sprintf("%s:%s", userAudienceCacheKey, userUUID)
	am.cacheClient.Del(cacheKey)
}

// InvalidateAllAudiencesCache removes the global audiences cache entry
func (am *AudienceManager) InvalidateAllAudiencesCache() {
	am.cacheClient.Del(audienceCacheKey)
}

func (am *AudienceManager) setSliceCache(key string, values []string) error {
	data, err := json.Marshal(values)
	if err != nil {
		return err
	}

	am.cacheClient.Set(key, string(data), 1)
	am.cacheClient.Wait()

	return nil
}

func (am *AudienceManager) getSliceCache(key string) ([]string, bool) {
	value, found := am.cacheClient.Get(key)
	if !found {
		return nil, false
	}

	var result []string
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		am.logger.Warnw("Failed to unmarshal cache value", "key", key, "error", err)

		return nil, false
	}

	return result, true
}
