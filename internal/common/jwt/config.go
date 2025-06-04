package jwt

import (
	"errors"

	"MydroX/anicetus/internal/config"
)

// Constants for security and validation
const (
	// Minimum length for JWT secret key
	minJWTSecretLength = 32

	// Maximum clock skew allowed in seconds
	minClockSkewSeconds = 0
	maxClockSkewSeconds = 300

	// Minimum and maximum durations for access tokens (in seconds)
	minAccessTokenDuration = 60
	maxAccessTokenDuration = 3600 // 1 hour

	// Minimum and maximum durations for refresh tokens (in seconds)
	minRefreshTokenDuration = 3600    // 1 hour
	maxRefreshTokenDuration = 2592000 // 30 days
)

// Constants for claims
const (
	claimTokenType   = "token_type"
	claimUserUUID    = "user_uuid"
	claimSessionUUID = "session_uuid"
	claimPermissions = "permissions"
	claimExp         = "exp"
	claimIAT         = "iat"
	claimIss         = "iss"
	claimAud         = "aud"
	claimNbf         = "nbf"
)

// TokenConfig holds configuration for the JWT service
type TokenConfig struct {
	// Secret key used for signing tokens
	SecretKey string

	// ExpectedIssuer is the expected issuer claim value
	ExpectedIssuer string

	// ExpectedAudiences is a list of valid audience values
	ExpectedAudiences []string

	// ClockSkewSeconds is the tolerance for time-based claims validation
	ClockSkewSeconds int

	// AccessTokenDuration is the default duration of access tokens in seconds
	AccessTokenDuration int

	// RefreshTokenDuration is the default duration of refresh tokens in seconds
	RefreshTokenDuration int
}

func NewTokenConfigFromEnv(cfg *config.Config) (TokenConfig, error) {
	// Validate JWT secret - mandatory for security
	if err := validateTokenConfig(cfg); err != nil {
		return TokenConfig{}, errors.New("invalid JWT configuration: " + err.Error())
	}

	// Create config with validated values
	tokenCfg := TokenConfig{
		SecretKey:            cfg.JWT.Secret,
		ExpectedIssuer:       cfg.JWT.Issuer,
		ExpectedAudiences:    []string{cfg.JWT.Issuer}, // Default to issuer as audience
		ClockSkewSeconds:     cfg.JWT.SkewSeconds,
		AccessTokenDuration:  cfg.JWT.AccessToken.Expiration,
		RefreshTokenDuration: cfg.JWT.RefreshToken.Expiration,
	}

	return tokenCfg, nil
}

func validateTokenConfig(cfg *config.Config) error {
	if cfg.JWT.Secret == "" {
		return errors.New("JWT secret is required in configuration")
	}

	// Validate secret key strength
	if len(cfg.JWT.Secret) < minJWTSecretLength {
		return errors.New("JWT secret must be at least 32 characters for security")
	}

	// Validate issuer - mandatory for proper JWT validation
	if cfg.JWT.Issuer == "" {
		return errors.New("JWT issuer is required in configuration")
	}

	// Security validation: Clock skew bounds
	if cfg.JWT.SkewSeconds < minClockSkewSeconds || cfg.JWT.SkewSeconds > maxClockSkewSeconds {
		return errors.New("JWT clock skew must be between 0 and 300 seconds")
	}

	// Security validation: Access token duration bounds
	if cfg.JWT.AccessToken.Expiration < minAccessTokenDuration || cfg.JWT.AccessToken.Expiration > maxAccessTokenDuration {
		return errors.New("JWT access token expiration must be between 60 and 3600 seconds")
	}

	// Security validation: Refresh token duration bounds
	if cfg.JWT.RefreshToken.Expiration < minRefreshTokenDuration || cfg.JWT.RefreshToken.Expiration > maxRefreshTokenDuration {
		return errors.New("JWT refresh token expiration must be between 3600 and 2592000 seconds")
	}

	return nil
}
