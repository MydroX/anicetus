package jwt

import (
	"testing"

	"MydroX/anicetus/internal/config"

	"github.com/stretchr/testify/assert"
)

const (
	testAccessSecret  = "test-access-secret-key-long-enough-for-jwt-signing"
	testRefreshSecret = "test-refresh-secret-key-long-enough-for-jwt-signin"
)

func validJWTConfig() config.JWT {
	return config.JWT{
		Issuer:      "test-issuer",
		SkewSeconds: 60,
		AccessToken: config.AccessToken{
			Secret:     testAccessSecret,
			Expiration: 900,
		},
		RefreshToken: config.RefreshToken{
			Secret:     testRefreshSecret,
			Expiration: 86400,
		},
	}
}

func TestNewTokenConfigFromEnv(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		cfg := &config.Config{JWT: validJWTConfig()}

		tokenCfg, err := NewTokenConfigFromEnv(cfg)
		assert.NoError(t, err)
		assert.Equal(t, testAccessSecret, tokenCfg.AccessTokenSecret)
		assert.Equal(t, testRefreshSecret, tokenCfg.RefreshTokenSecret)
		assert.Equal(t, "test-issuer", tokenCfg.ExpectedIssuer)
		assert.Equal(t, []string{"test-issuer"}, tokenCfg.ExpectedAudiences)
		assert.Equal(t, 60, tokenCfg.ClockSkewSeconds)
		assert.Equal(t, 900, tokenCfg.AccessTokenDuration)
		assert.Equal(t, 86400, tokenCfg.RefreshTokenDuration)
	})

	t.Run("invalid configuration", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Issuer: "test-issuer",
				AccessToken: config.AccessToken{
					Secret: "short",
				},
			},
		}

		_, err := NewTokenConfigFromEnv(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JWT configuration")
	})
}

func TestValidateTokenConfig(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		cfg := &config.Config{JWT: validJWTConfig()}
		assert.NoError(t, validateTokenConfig(cfg))
	})

	t.Run("missing access token secret", func(t *testing.T) {
		cfg := &config.Config{JWT: validJWTConfig()}
		cfg.JWT.AccessToken.Secret = ""
		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access token secret is required")
	})

	t.Run("access token secret too short", func(t *testing.T) {
		cfg := &config.Config{JWT: validJWTConfig()}
		cfg.JWT.AccessToken.Secret = "short"
		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access token secret must be at least 32 characters")
	})

	t.Run("missing refresh token secret", func(t *testing.T) {
		cfg := &config.Config{JWT: validJWTConfig()}
		cfg.JWT.RefreshToken.Secret = ""
		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "refresh token secret is required")
	})

	t.Run("refresh token secret too short", func(t *testing.T) {
		cfg := &config.Config{JWT: validJWTConfig()}
		cfg.JWT.RefreshToken.Secret = "short"
		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "refresh token secret must be at least 32 characters")
	})

	t.Run("same secret for access and refresh", func(t *testing.T) {
		cfg := &config.Config{JWT: validJWTConfig()}
		cfg.JWT.RefreshToken.Secret = cfg.JWT.AccessToken.Secret
		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be different")
	})

	t.Run("missing issuer", func(t *testing.T) {
		cfg := &config.Config{JWT: validJWTConfig()}
		cfg.JWT.Issuer = ""
		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "issuer is required")
	})

	t.Run("clock skew out of range", func(t *testing.T) {
		cfg := &config.Config{JWT: validJWTConfig()}
		cfg.JWT.SkewSeconds = -1
		assert.Error(t, validateTokenConfig(cfg))

		cfg.JWT.SkewSeconds = 301
		assert.Error(t, validateTokenConfig(cfg))
	})

	t.Run("clock skew boundary values", func(t *testing.T) {
		cfg := &config.Config{JWT: validJWTConfig()}
		cfg.JWT.SkewSeconds = 0
		assert.NoError(t, validateTokenConfig(cfg))

		cfg.JWT.SkewSeconds = 300
		assert.NoError(t, validateTokenConfig(cfg))
	})

	t.Run("access token duration out of range", func(t *testing.T) {
		cfg := &config.Config{JWT: validJWTConfig()}
		cfg.JWT.AccessToken.Expiration = 59
		assert.Error(t, validateTokenConfig(cfg))

		cfg.JWT.AccessToken.Expiration = 3601
		assert.Error(t, validateTokenConfig(cfg))
	})

	t.Run("access token duration boundary values", func(t *testing.T) {
		cfg := &config.Config{JWT: validJWTConfig()}
		cfg.JWT.AccessToken.Expiration = 60
		assert.NoError(t, validateTokenConfig(cfg))

		cfg.JWT.AccessToken.Expiration = 3600
		assert.NoError(t, validateTokenConfig(cfg))
	})

	t.Run("refresh token duration out of range", func(t *testing.T) {
		cfg := &config.Config{JWT: validJWTConfig()}
		cfg.JWT.RefreshToken.Expiration = 3599
		assert.Error(t, validateTokenConfig(cfg))

		cfg.JWT.RefreshToken.Expiration = 2592001
		assert.Error(t, validateTokenConfig(cfg))
	})

	t.Run("refresh token duration boundary values", func(t *testing.T) {
		cfg := &config.Config{JWT: validJWTConfig()}
		cfg.JWT.RefreshToken.Expiration = 3600
		assert.NoError(t, validateTokenConfig(cfg))

		cfg.JWT.RefreshToken.Expiration = 2592000
		assert.NoError(t, validateTokenConfig(cfg))
	})
}
