package usecases

import (
	"context"
	"encoding/json"
	"fmt"

	iamrepository "MydroX/anicetus/internal/iam/repository"

	"github.com/valkey-io/valkey-go"
	"go.uber.org/zap"
)

const (
	audienceCacheKey     = "allowed_audiences"
	userAudienceCacheKey = "user_audiences"
)

// AudienceManager handles caching and retrieval of allowed JWT audiences
type AudienceManager struct {
	logger     *zap.SugaredLogger
	repository iamrepository.AudienceStore
	valkey     valkey.Client
}

// NewAudienceManager creates a new audience manager
func NewAudienceManager(logger *zap.SugaredLogger, repo iamrepository.AudienceStore, client valkey.Client) *AudienceManager {
	return &AudienceManager{
		logger:     logger,
		repository: repo,
		valkey:     client,
	}
}

// CacheAllowedAudiences loads audiences from database into cache
func (am *AudienceManager) CacheAllowedAudiences(ctx context.Context) error {
	audiences, err := am.repository.GetAllowedAudiences(ctx)
	if err != nil {
		return err
	}

	return am.setSliceCache(ctx, audienceCacheKey, audiences)
}

// GetAllowedAudiences retrieves audiences from cache or database
func (am *AudienceManager) GetAllowedAudiences(ctx context.Context) ([]string, error) {
	if audiences, ok := am.getSliceCache(ctx, audienceCacheKey); ok {
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

	if audiences, ok := am.getSliceCache(ctx, cacheKey); ok {
		return audiences, nil
	}

	audiences, err := am.repository.GetUserAudiences(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	_ = am.setSliceCache(ctx, cacheKey, audiences)

	return audiences, nil
}

// InvalidateUserAudiencesCache removes a user's audience cache entry
func (am *AudienceManager) InvalidateUserAudiencesCache(ctx context.Context, userUUID string) {
	cacheKey := fmt.Sprintf("%s:%s", userAudienceCacheKey, userUUID)

	am.valkey.Do(ctx, am.valkey.B().Del().Key(cacheKey).Build())
}

// InvalidateAllAudiencesCache removes the global audiences cache entry
func (am *AudienceManager) InvalidateAllAudiencesCache(ctx context.Context) {
	am.valkey.Do(ctx, am.valkey.B().Del().Key(audienceCacheKey).Build())
}

func (am *AudienceManager) setSliceCache(ctx context.Context, key string, values []string) error {
	data, err := json.Marshal(values)
	if err != nil {
		return err
	}

	return am.valkey.Do(ctx, am.valkey.B().Set().Key(key).Value(string(data)).Build()).Error()
}

func (am *AudienceManager) getSliceCache(ctx context.Context, key string) ([]string, bool) {
	result, err := am.valkey.Do(ctx, am.valkey.B().Get().Key(key).Build()).ToString()
	if err != nil {
		return nil, false
	}

	var audiences []string
	if err := json.Unmarshal([]byte(result), &audiences); err != nil {
		am.logger.Warnw("Failed to unmarshal cache value", "key", key, "error", err)

		return nil, false
	}

	return audiences, true
}
