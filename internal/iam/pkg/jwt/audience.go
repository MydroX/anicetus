// audience.go
package jwt

import (
	"encoding/json"

	iamrepository "MydroX/anicetus/internal/iam/repository"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	audienceCacheKey = "allowed_audience"
)

// AudienceManager handles caching and retrieval of allowed JWT audiences
type AudienceManager struct {
	logger      *zap.SugaredLogger
	db          *pgxpool.Pool
	cacheClient *ristretto.Cache[string, string]
}

// NewAudienceManager creates a new audience manager
func NewAudienceManager(logger *zap.SugaredLogger, db *pgxpool.Pool, cache *ristretto.Cache[string, string]) *AudienceManager {
	return &AudienceManager{
		logger:      logger,
		db:          db,
		cacheClient: cache,
	}
}

// CacheAllowedAudiences loads audiences from database into cache
func (am *AudienceManager) CacheAllowedAudiences() error {
	// Get audiences from repository
	repository := iamrepository.NewAudienceStore(am.logger, am.db)

	audiences, err := repository.GetAllowedAudiences()
	if err != nil {
		return err
	}

	// Convert the slice of audiences to a single string value
	audiencesJSON, err := json.Marshal(audiences)
	if err != nil {
		return err
	}

	// Store in cache without expiration
	success := am.cacheClient.Set(audienceCacheKey, string(audiencesJSON), 1)
	if !success {
		am.logger.Warn("Failed to set audiences in cache")
	}

	// Force write to storage - Ristretto uses a buffer
	am.cacheClient.Wait()

	return nil
}

// GetAllowedAudiences retrieves audiences from cache or database
func (am *AudienceManager) GetAllowedAudiences() ([]string, error) {
	// Try to get from cache first
	value, found := am.cacheClient.Get(audienceCacheKey)
	if found {
		// Cache hit - unmarshal the value
		var audiences []string

		err := json.Unmarshal([]byte(value), &audiences)
		if err == nil {
			return audiences, nil
		}

		am.logger.Warn("Failed to unmarshal audiences from cache", "error", err.Error())
	}

	// Cache miss or unmarshal error - get from repository and update cache
	repository := iamrepository.NewAudienceStore(am.logger, am.db)

	audiences, err := repository.GetAllowedAudiences()
	if err != nil {
		return nil, err
	}

	// Update cache in background to not block the current request
	go func() {
		if err := am.CacheAllowedAudiences(); err != nil {
			am.logger.Warn("Failed to update audiences cache", "error", err.Error())
		}
	}()

	return audiences, nil
}

// UpdateAllowedAudiencesCache triggers a manual cache update
func (am *AudienceManager) UpdateAllowedAudiencesCache() error {
	return am.CacheAllowedAudiences()
}
