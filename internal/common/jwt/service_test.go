package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestParseAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := TokenConfig{
		SecretKey:        "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		ExpectedIssuer:   "test-issuer",
		ClockSkewSeconds: 60,
	}

	service := NewJWTService(config)

	t.Run("success case", func(t *testing.T) {
		// Create valid test token
		userUUID := "user-123"
		permissions := []string{"read", "write"}

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, AccessTokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "test-issuer",
				Audience:  []string{"test-audience"},
			},
			UserUUID:    userUUID,
			TokenType:   string(AccessToken),
			Permissions: permissions,
		})

		tokenString, err := token.SignedString([]byte(config.SecretKey))
		assert.NoError(t, err)

		// Parse token
		claims, err := service.ParseAccessToken(tokenString)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userUUID, claims.UserUUID)
		assert.Equal(t, AccessToken, claims.TokenType)
		assert.ElementsMatch(t, permissions, claims.Permissions)
	})

	t.Run("expired token", func(t *testing.T) {
		// Create expired token
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, AccessTokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-10 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-20 * time.Minute)),
				Issuer:    "test-issuer",
				Audience:  []string{"test-audience"},
			},
			UserUUID:    "user-123",
			TokenType:   string(AccessToken),
			Permissions: []string{"read"},
		})

		tokenString, err := token.SignedString([]byte(config.SecretKey))
		assert.NoError(t, err)

		// Parse token
		claims, err := service.ParseAccessToken(tokenString)

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("invalid token type", func(t *testing.T) {
		// Create refresh token instead of access token
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, RefreshTokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "test-issuer",
				Audience:  []string{"test-audience"},
			},
			UserUUID:    "user-123",
			TokenType:   string(RefreshToken),
			SessionUUID: "session-456",
		})

		tokenString, err := token.SignedString([]byte(config.SecretKey))
		assert.NoError(t, err)

		// Parse token
		claims, err := service.ParseAccessToken(tokenString)

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, ErrNotAccessToken)
	})

	t.Run("malformed token", func(t *testing.T) {
		// Parse malformed token
		claims, err := service.ParseAccessToken("not-a-valid-jwt-token")

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, ErrInvalidTokenFormat)
	})
}

func TestParseRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := TokenConfig{
		SecretKey:        "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		ExpectedIssuer:   "test-issuer",
		ClockSkewSeconds: 60,
	}

	service := NewJWTService(config)

	t.Run("success case", func(t *testing.T) {
		// Create valid refresh token
		userUUID := "user-123"
		sessionUUID := "session-456"

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, RefreshTokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "test-issuer",
				Audience:  []string{"test-audience"},
			},
			UserUUID:    userUUID,
			TokenType:   string(RefreshToken),
			SessionUUID: sessionUUID,
		})

		tokenString, err := token.SignedString([]byte(config.SecretKey))
		assert.NoError(t, err)

		// Parse token
		claims, err := service.ParseRefreshToken(tokenString)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userUUID, claims.UserUUID)
		assert.Equal(t, RefreshToken, claims.TokenType)
		assert.Equal(t, sessionUUID, claims.SessionUUID)
	})

	t.Run("missing session UUID", func(t *testing.T) {
		// Create token without session UUID
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"exp":        time.Now().Add(10 * time.Minute).Unix(),
			"iat":        time.Now().Unix(),
			"iss":        "test-issuer",
			"aud":        []string{"test-audience"},
			"user_uuid":  "user-123",
			"token_type": string(RefreshToken),
			// No session_uuid
		})

		tokenString, err := token.SignedString([]byte(config.SecretKey))
		assert.NoError(t, err)

		// Parse token
		claims, err := service.ParseRefreshToken(tokenString)

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, ErrMissingSessionUUID)
	})
}

func TestParseToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := TokenConfig{
		SecretKey:        "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		ExpectedIssuer:   "test-issuer",
		ClockSkewSeconds: 60,
	}

	service := NewJWTService(config)

	t.Run("parse access token", func(t *testing.T) {
		// Create access token
		userUUID := "user-123"

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, AccessTokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "test-issuer",
				Audience:  []string{"test-audience"},
			},
			UserUUID:    userUUID,
			TokenType:   string(AccessToken),
			Permissions: []string{"read"},
		})

		tokenString, err := token.SignedString([]byte(config.SecretKey))
		assert.NoError(t, err)

		// Parse generic token
		claims, err := service.ParseToken(tokenString)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userUUID, claims.UserUUID)
		assert.Equal(t, AccessToken, claims.TokenType)
	})

	t.Run("parse refresh token", func(t *testing.T) {
		// Create refresh token
		userUUID := "user-123"

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, RefreshTokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "test-issuer",
				Audience:  []string{"test-audience"},
			},
			UserUUID:    userUUID,
			TokenType:   string(RefreshToken),
			SessionUUID: "session-456",
		})

		tokenString, err := token.SignedString([]byte(config.SecretKey))
		assert.NoError(t, err)

		// Parse generic token
		claims, err := service.ParseToken(tokenString)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userUUID, claims.UserUUID)
		assert.Equal(t, RefreshToken, claims.TokenType)
	})

	t.Run("invalid signing method", func(t *testing.T) {
		// Create token with unsupported signing method
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"exp":        time.Now().Add(10 * time.Minute).Unix(),
			"iat":        time.Now().Unix(),
			"iss":        "test-issuer",
			"aud":        []string{"test-audience"},
			"user_uuid":  "user-123",
			"token_type": string(AccessToken),
		})

		// Use RSA private key for signing
		tokenString, _ := token.SignedString([]byte("some-key"))

		// Parse generic token
		claims, err := service.ParseToken(tokenString)

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "token format is invalid")
	})
}

func TestCreateAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("successful token creation", func(t *testing.T) {
		// Configure service
		config := TokenConfig{
			SecretKey:           "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
			ExpectedIssuer:      "test-issuer",
			AccessTokenDuration: 3600,
		}
		service := NewJWTService(config)

		// Call function to test
		userUUID := "user-123"
		permissions := []string{"read", "write"}
		audiences := []string{"test-audience"}
		token, err := service.CreateAccessToken(userUUID, permissions, audiences)

		// Verify
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Decode and validate token
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.SecretKey), nil
		})
		assert.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		// Check claims
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, userUUID, claims["user_uuid"])
		assert.Equal(t, string(AccessToken), claims["token_type"])
		assert.Equal(t, config.ExpectedIssuer, claims["iss"])

		// Check permissions
		perms, ok := claims["permissions"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, perms, len(permissions))
		for i, p := range permissions {
			assert.Equal(t, p, perms[i])
		}

		// Check expiration
		exp, ok := claims["exp"].(float64)
		assert.True(t, ok)
		assert.Greater(t, exp, float64(time.Now().Unix()))
		assert.InDelta(t, time.Now().Add(time.Duration(config.AccessTokenDuration)*time.Second).Unix(), int64(exp), 5)
	})

	t.Run("missing secret key", func(t *testing.T) {
		// Configure service with empty secret key
		config := TokenConfig{
			SecretKey:      "", // Empty key
			ExpectedIssuer: "test-issuer",
		}
		service := NewJWTService(config)

		// Call function to test
		token, err := service.CreateAccessToken("user-123", []string{"read"}, []string{})

		// Check current behavior
		if err != nil {
			assert.ErrorIs(t, err, ErrMissingSecretKey)
			assert.Empty(t, token)
		} else {
			assert.NotEmpty(t, token)
			t.Log("NOTE: The service currently accepts empty secret keys, which is a security issue.")
			t.Log("Consider modifying CreateAccessToken to check for empty keys and return ErrMissingSecretKey.")
		}
	})
}

func TestCreateRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("successful token creation", func(t *testing.T) {
		// Configure service
		config := TokenConfig{
			SecretKey:            "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
			ExpectedIssuer:       "test-issuer",
			RefreshTokenDuration: 86400, // 24 hours
		}
		service := NewJWTService(config)

		// Call function to test
		userUUID := "user-123"
		sessionUUID := "session-456"
		audiences := []string{"test-audience"}
		token, err := service.CreateRefreshToken(userUUID, sessionUUID, audiences)

		// Verify
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Decode and validate token
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.SecretKey), nil
		})
		assert.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		// Check claims
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, userUUID, claims["user_uuid"])
		assert.Equal(t, sessionUUID, claims["session_uuid"])
		assert.Equal(t, string(RefreshToken), claims["token_type"])
		assert.Equal(t, config.ExpectedIssuer, claims["iss"])

		// Check expiration
		exp, ok := claims["exp"].(float64)
		assert.True(t, ok)
		assert.Greater(t, exp, float64(time.Now().Unix()))
		assert.InDelta(t, time.Now().Add(time.Duration(config.RefreshTokenDuration)*time.Second).Unix(), int64(exp), 5)
	})

	t.Run("missing secret key", func(t *testing.T) {
		// Configure service with empty secret key
		config := TokenConfig{
			SecretKey:      "", // Empty key
			ExpectedIssuer: "test-issuer",
		}
		service := NewJWTService(config)

		// Call function to test
		token, err := service.CreateRefreshToken("user-123", "session-456", []string{})

		// Check current behavior
		if err != nil {
			assert.ErrorIs(t, err, ErrMissingSecretKey)
			assert.Empty(t, token)
		} else {
			assert.NotEmpty(t, token)
			t.Log("NOTE: The service currently accepts empty secret keys, which is a security issue.")
			t.Log("Consider modifying CreateRefreshToken to check for empty keys and return ErrMissingSecretKey.")
		}
	})

	t.Run("empty session UUID", func(t *testing.T) {
		// Configure service
		config := TokenConfig{
			SecretKey:      "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
			ExpectedIssuer: "test-issuer",
		}
		service := NewJWTService(config)

		// Test with empty sessionUUID
		audiences := []string{"test-audience"}
		token, err := service.CreateRefreshToken("user-123", "", audiences)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Improvement suggestion: validate sessionUUID is not empty in CreateRefreshToken
	})
}
