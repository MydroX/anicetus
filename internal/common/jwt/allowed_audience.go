package jwt

import (
	iamrepository "MydroX/anicetus/internal/iam/repository"
	"encoding/json"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	audienceCacheKey = "allowed_audience"
)

func AllowedAudiencesInCache(db *pgxpool.Pool, cacheClient *ristretto.Cache[string, string], l *zap.SugaredLogger) error {
	// Get audiences from repository
	repository := iamrepository.NewAudienceStore(l, db)
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
	success := cacheClient.Set(audienceCacheKey, string(audiencesJSON), 1)
	if !success {
		l.Warn("Failed to set audiences in cache")
	}

	// Force write to storage - Ristretto uses a buffer
	cacheClient.Wait()

	return nil
}

func GetAllowedAudiences(cacheClient *ristretto.Cache[string, string], db *pgxpool.Pool, l *zap.SugaredLogger) ([]string, error) {
	// Try to get from cache first
	value, found := cacheClient.Get(audienceCacheKey)
	if found {
		// Cache hit - unmarshal the value
		var audiences []string
		err := json.Unmarshal([]byte(value), &audiences)
		if err == nil {
			return audiences, nil
		}

		l.Warn("Failed to unmarshal audiences from cache", "error", err.Error())
	}

	// Cache miss or unmarshal error - get from repository and update cache
	repository := iamrepository.NewAudienceStore(l, db)
	audiences, err := repository.GetAllowedAudiences()
	if err != nil {
		return nil, err
	}

	// Update cache for future requests
	go func() {
		// Update cache in background to not block the current request
		if err := AllowedAudiencesInCache(db, cacheClient, l); err != nil {
			l.Warn("Failed to update audiences cache", "error", err.Error())
		}
	}()

	return audiences, nil
}
