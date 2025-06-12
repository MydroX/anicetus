package jwt

import (
	"testing"

	"MydroX/anicetus/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestNewTokenConfigFromEnv(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "this_is_a_very_secure_32_char_secret_key_for_testing_purposes",
				Issuer:      "test-issuer",
				SkewSeconds: 60,
				AccessToken: config.AccessToken{
					Expiration: 900, // 15 minutes
				},
				RefreshToken: config.RefreshToken{
					Expiration: 86400, // 24 hours
				},
			},
		}

		tokenCfg, err := NewTokenConfigFromEnv(cfg)
		assert.NoError(t, err)
		assert.Equal(t, cfg.JWT.Secret, tokenCfg.SecretKey)
		assert.Equal(t, cfg.JWT.Issuer, tokenCfg.ExpectedIssuer)
		assert.Equal(t, []string{cfg.JWT.Issuer}, tokenCfg.ExpectedAudiences)
		assert.Equal(t, cfg.JWT.SkewSeconds, tokenCfg.ClockSkewSeconds)
		assert.Equal(t, cfg.JWT.AccessToken.Expiration, tokenCfg.AccessTokenDuration)
		assert.Equal(t, cfg.JWT.RefreshToken.Expiration, tokenCfg.RefreshTokenDuration)
	})

	t.Run("invalid configuration", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret: "short", // Too short
				Issuer: "test-issuer",
			},
		}

		_, err := NewTokenConfigFromEnv(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JWT configuration")
	})
}

func TestValidateTokenConfig(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "this_is_a_very_secure_32_char_secret_key_for_testing_purposes",
				Issuer:      "test-issuer",
				SkewSeconds: 60,
				AccessToken: config.AccessToken{
					Expiration: 900,
				},
				RefreshToken: config.RefreshToken{
					Expiration: 86400,
				},
			},
		}

		err := validateTokenConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("missing JWT secret", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret: "",
				Issuer: "test-issuer",
			},
		}

		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT secret is required")
	})

	t.Run("JWT secret too short", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret: "short_secret", // Only 12 characters
				Issuer: "test-issuer",
			},
		}

		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT secret must be at least 32 characters")
	})

	t.Run("JWT secret exactly minimum length", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "12345678901234567890123456789012", // Exactly 32 characters
				Issuer:      "test-issuer",
				SkewSeconds: 60,
				AccessToken: config.AccessToken{
					Expiration: 900,
				},
				RefreshToken: config.RefreshToken{
					Expiration: 86400,
				},
			},
		}

		err := validateTokenConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("missing JWT issuer", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret: "this_is_a_very_secure_32_char_secret_key_for_testing_purposes",
				Issuer: "",
			},
		}

		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT issuer is required")
	})

	t.Run("clock skew too low", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "this_is_a_very_secure_32_char_secret_key_for_testing_purposes",
				Issuer:      "test-issuer",
				SkewSeconds: -1, // Below minimum
			},
		}

		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT clock skew must be between 0 and 300 seconds")
	})

	t.Run("clock skew too high", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "this_is_a_very_secure_32_char_secret_key_for_testing_purposes",
				Issuer:      "test-issuer",
				SkewSeconds: 301, // Above maximum
			},
		}

		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT clock skew must be between 0 and 300 seconds")
	})

	t.Run("clock skew boundary values", func(t *testing.T) {
		// Test minimum boundary (0)
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "this_is_a_very_secure_32_char_secret_key_for_testing_purposes",
				Issuer:      "test-issuer",
				SkewSeconds: 0,
				AccessToken: config.AccessToken{
					Expiration: 900,
				},
				RefreshToken: config.RefreshToken{
					Expiration: 86400,
				},
			},
		}

		err := validateTokenConfig(cfg)
		assert.NoError(t, err)

		// Test maximum boundary (300)
		cfg.JWT.SkewSeconds = 300
		err = validateTokenConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("access token duration too short", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "this_is_a_very_secure_32_char_secret_key_for_testing_purposes",
				Issuer:      "test-issuer",
				SkewSeconds: 60,
				AccessToken: config.AccessToken{
					Expiration: 59, // Below minimum
				},
			},
		}

		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT access token expiration must be between 60 and 3600 seconds")
	})

	t.Run("access token duration too long", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "this_is_a_very_secure_32_char_secret_key_for_testing_purposes",
				Issuer:      "test-issuer",
				SkewSeconds: 60,
				AccessToken: config.AccessToken{
					Expiration: 3601, // Above maximum
				},
			},
		}

		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT access token expiration must be between 60 and 3600 seconds")
	})

	t.Run("access token boundary values", func(t *testing.T) {
		// Test minimum boundary (60)
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "this_is_a_very_secure_32_char_secret_key_for_testing_purposes",
				Issuer:      "test-issuer",
				SkewSeconds: 60,
				AccessToken: config.AccessToken{
					Expiration: 60,
				},
				RefreshToken: config.RefreshToken{
					Expiration: 86400,
				},
			},
		}

		err := validateTokenConfig(cfg)
		assert.NoError(t, err)

		// Test maximum boundary (3600)
		cfg.JWT.AccessToken.Expiration = 3600
		err = validateTokenConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("refresh token duration too short", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "this_is_a_very_secure_32_char_secret_key_for_testing_purposes",
				Issuer:      "test-issuer",
				SkewSeconds: 60,
				AccessToken: config.AccessToken{
					Expiration: 900,
				},
				RefreshToken: config.RefreshToken{
					Expiration: 3599, // Below minimum
				},
			},
		}

		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT refresh token expiration must be between 3600 and 2592000 seconds")
	})

	t.Run("refresh token duration too long", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "this_is_a_very_secure_32_char_secret_key_for_testing_purposes",
				Issuer:      "test-issuer",
				SkewSeconds: 60,
				AccessToken: config.AccessToken{
					Expiration: 900,
				},
				RefreshToken: config.RefreshToken{
					Expiration: 2592001, // Above maximum
				},
			},
		}

		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT refresh token expiration must be between 3600 and 2592000 seconds")
	})

	t.Run("refresh token boundary values", func(t *testing.T) {
		// Test minimum boundary (3600)
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "this_is_a_very_secure_32_char_secret_key_for_testing_purposes",
				Issuer:      "test-issuer",
				SkewSeconds: 60,
				AccessToken: config.AccessToken{
					Expiration: 900,
				},
				RefreshToken: config.RefreshToken{
					Expiration: 3600,
				},
			},
		}

		err := validateTokenConfig(cfg)
		assert.NoError(t, err)

		// Test maximum boundary (2592000)
		cfg.JWT.RefreshToken.Expiration = 2592000
		err = validateTokenConfig(cfg)
		assert.NoError(t, err)
	})
}

func TestSecurityEdgeCasesConfig(t *testing.T) {
	t.Run("weak secret exactly at boundary", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret: "1234567890123456789012345678901", // 31 characters - just below minimum
				Issuer: "test-issuer",
			},
		}

		err := validateTokenConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT secret must be at least 32 characters")
	})

	t.Run("very long but valid secret", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "this_is_a_very_very_very_long_secret_key_that_should_still_be_valid_for_security_purposes_and_testing",
				Issuer:      "test-issuer",
				SkewSeconds: 60,
				AccessToken: config.AccessToken{
					Expiration: 900,
				},
				RefreshToken: config.RefreshToken{
					Expiration: 86400,
				},
			},
		}

		err := validateTokenConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("zero clock skew security implications", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "this_is_a_very_secure_32_char_secret_key_for_testing_purposes",
				Issuer:      "test-issuer",
				SkewSeconds: 0, // No tolerance - strict timing
				AccessToken: config.AccessToken{
					Expiration: 900,
				},
				RefreshToken: config.RefreshToken{
					Expiration: 86400,
				},
			},
		}

		err := validateTokenConfig(cfg)
		assert.NoError(t, err, "Zero clock skew should be valid for high security environments")
	})

	t.Run("maximum clock skew security risk", func(t *testing.T) {
		cfg := &config.Config{
			JWT: config.JWT{
				Secret:      "this_is_a_very_secure_32_char_secret_key_for_testing_purposes",
				Issuer:      "test-issuer",
				SkewSeconds: 300, // Maximum allowed - 5 minutes tolerance
				AccessToken: config.AccessToken{
					Expiration: 900,
				},
				RefreshToken: config.RefreshToken{
					Expiration: 86400,
				},
			},
		}

		err := validateTokenConfig(cfg)
		assert.NoError(t, err, "Maximum clock skew should be valid but represents higher security risk")
	})
}
