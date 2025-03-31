//revive:disable:add-constant
package jwt

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSecretKey   = "test-secret-key-for-jwt-testing-purposes-only"
	testUserUUID    = "test-user-uuid-123"
	testSessionUUID = "test-session-uuid-456"
)

func setupTestEnv() func() {
	originalSecret := os.Getenv("JWT_SECRET")
	err := os.Setenv("JWT_SECRET", testSecretKey)
	if err != nil {
		panic("Failed to set JWT_SECRET environment variable")
	}

	return func() {
		err = os.Setenv("JWT_SECRET", originalSecret)
		if err != nil {
			panic("Failed to restore JWT_SECRET environment variable")
		}
	}
}

func TestCreateAccessToken(t *testing.T) {
	tests := []struct {
		name        string
		claims      *AccessClaims
		secretKey   string
		expectError bool
		expectedErr error
	}{
		{
			name: "Valid access token",
			claims: &AccessClaims{
				BaseClaims: BaseClaims{
					UserUUID:           testUserUUID,
					TokenType:          AccessToken,
					ExpirationDuration: time.Hour,
				},
				Permissions: []string{"read", "write"},
			},
			secretKey:   testSecretKey,
			expectError: false,
		},
		{
			name: "Missing secret key",
			claims: &AccessClaims{
				BaseClaims: BaseClaims{
					UserUUID:           testUserUUID,
					TokenType:          AccessToken,
					ExpirationDuration: time.Hour,
				},
			},
			secretKey:   "",
			expectError: true,
			expectedErr: ErrMissingSecretKey,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, err := CreateAccessToken(tc.claims, tc.secretKey)

			if tc.expectError {
				assert.Error(t, err)
				var jwtErr *JWTError
				assert.True(t, errors.As(err, &jwtErr))
				assert.True(t, errors.Is(jwtErr, tc.expectedErr))
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify token content
				parsedToken, err := jwt.Parse(token, func(_ *jwt.Token) (any, error) {
					return []byte(tc.secretKey), nil
				})
				require.NoError(t, err)
				assert.True(t, parsedToken.Valid)

				claims, ok := parsedToken.Claims.(jwt.MapClaims)
				require.True(t, ok)

				assert.Equal(t, testUserUUID, claims["user_uuid"])
				assert.Equal(t, string(AccessToken), claims["token_type"])
				assert.NotNil(t, claims["exp"])
				assert.NotNil(t, claims["iat"])

				// Check permissions
				permissions, ok := claims["permissions"].([]any)
				require.True(t, ok)
				assert.Len(t, permissions, len(tc.claims.Permissions))
			}
		})
	}
}

func TestCreateRefreshToken(t *testing.T) {
	tests := []struct {
		name        string
		claims      *RefreshClaims
		secretKey   string
		expectError bool
		expectedErr error
	}{
		{
			name: "Valid refresh token",
			claims: &RefreshClaims{
				BaseClaims: BaseClaims{
					UserUUID:           testUserUUID,
					TokenType:          RefreshToken,
					ExpirationDuration: time.Hour * 24 * 7, // 7 days
				},
				SessionUUID: testSessionUUID,
			},
			secretKey:   testSecretKey,
			expectError: false,
		},
		{
			name: "Missing secret key",
			claims: &RefreshClaims{
				BaseClaims: BaseClaims{
					UserUUID:           testUserUUID,
					TokenType:          RefreshToken,
					ExpirationDuration: time.Hour * 24 * 7,
				},
				SessionUUID: testSessionUUID,
			},
			secretKey:   "",
			expectError: true,
			expectedErr: ErrMissingSecretKey,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, err := CreateRefreshToken(tc.claims, tc.secretKey)

			if tc.expectError {
				assert.Error(t, err)
				var jwtErr *JWTError
				assert.True(t, errors.As(err, &jwtErr))
				assert.True(t, errors.Is(jwtErr, tc.expectedErr))
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify token content
				parsedToken, err := jwt.Parse(token, func(_ *jwt.Token) (any, error) {
					return []byte(tc.secretKey), nil
				})
				require.NoError(t, err)
				assert.True(t, parsedToken.Valid)

				claims, ok := parsedToken.Claims.(jwt.MapClaims)
				require.True(t, ok)

				assert.Equal(t, testUserUUID, claims["user_uuid"])
				assert.Equal(t, testSessionUUID, claims["session_uuid"])
				assert.Equal(t, string(RefreshToken), claims["token_type"])
				assert.NotNil(t, claims["exp"])
				assert.NotNil(t, claims["iat"])
			}
		})
	}
}

func TestParseAccessToken(t *testing.T) {
	cleanup := setupTestEnv()
	defer cleanup()

	validClaims := &AccessClaims{
		BaseClaims: BaseClaims{
			UserUUID:           testUserUUID,
			TokenType:          AccessToken,
			ExpirationDuration: time.Hour,
		},
		Permissions: []string{"read", "write"},
	}

	validToken, err := CreateAccessToken(validClaims, testSecretKey)
	require.NoError(t, err)

	// Create an expired token
	expiredClaims := &AccessClaims{
		BaseClaims: BaseClaims{
			UserUUID:           testUserUUID,
			TokenType:          AccessToken,
			ExpirationDuration: -time.Hour, // Expired 1 hour ago
		},
		Permissions: []string{"read"},
	}
	expiredToken, err := CreateAccessToken(expiredClaims, testSecretKey)
	require.NoError(t, err)

	// Create a refresh token (wrong type)
	refreshClaims := &RefreshClaims{
		BaseClaims: BaseClaims{
			UserUUID:           testUserUUID,
			TokenType:          RefreshToken,
			ExpirationDuration: time.Hour,
		},
		SessionUUID: testSessionUUID,
	}
	refreshToken, err := CreateRefreshToken(refreshClaims, testSecretKey)
	require.NoError(t, err)

	tests := []struct {
		name        string
		tokenString string
		expectError bool
		expectedErr error
	}{
		{
			name:        "Valid access token",
			tokenString: validToken,
			expectError: false,
		},
		{
			name:        "Expired token",
			tokenString: expiredToken,
			expectError: true,
			expectedErr: ErrTokenExpired,
		},
		{
			name:        "Not an access token",
			tokenString: refreshToken,
			expectError: true,
			expectedErr: ErrNotAccessToken,
		},
		{
			name:        "Invalid token format",
			tokenString: "invalid.token.format",
			expectError: true,
			expectedErr: ErrInvalidTokenFormat,
		},
		{
			name:        "Empty token",
			tokenString: "",
			expectError: true,
			expectedErr: ErrInvalidTokenFormat,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := ParseAccessToken(tc.tokenString)

			if tc.expectError {
				assert.Error(t, err)
				var jwtErr *JWTError
				assert.True(t, errors.As(err, &jwtErr), "Expected JWTError type")
				assert.True(t, errors.Is(jwtErr, tc.expectedErr),
					"Expected %v, got %v", tc.expectedErr, jwtErr)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, testUserUUID, claims.UserUUID)
				assert.Equal(t, AccessToken, claims.TokenType)
				assert.ElementsMatch(t, validClaims.Permissions, claims.Permissions)
			}
		})
	}
}

func TestParseRefreshToken(t *testing.T) {
	cleanup := setupTestEnv()
	defer cleanup()

	validClaims := &RefreshClaims{
		BaseClaims: BaseClaims{
			UserUUID:           testUserUUID,
			TokenType:          RefreshToken,
			ExpirationDuration: time.Hour,
		},
		SessionUUID: testSessionUUID,
	}

	validToken, err := CreateRefreshToken(validClaims, testSecretKey)
	require.NoError(t, err)

	_ = validToken

	// Create an expired token
	expiredClaims := &RefreshClaims{
		BaseClaims: BaseClaims{
			UserUUID:           testUserUUID,
			TokenType:          RefreshToken,
			ExpirationDuration: -time.Hour, // Expired 1 hour ago
		},
		SessionUUID: testSessionUUID,
	}
	expiredToken, err := CreateRefreshToken(expiredClaims, testSecretKey)
	require.NoError(t, err)

	_ = expiredToken

	// Create an access token (wrong type)
	accessClaims := &AccessClaims{
		BaseClaims: BaseClaims{
			UserUUID:           testUserUUID,
			TokenType:          AccessToken,
			ExpirationDuration: time.Hour,
		},
		Permissions: []string{"read"},
	}
	accessToken, err := CreateAccessToken(accessClaims, testSecretKey)
	require.NoError(t, err)

	_ = accessToken

	tests := []struct {
		name        string
		tokenString string
		expectError bool
		expectedErr error
	}{
		{
			name:        "Valid refresh token",
			tokenString: validToken,
			expectError: false,
		},
		{
			name:        "Expired token",
			tokenString: expiredToken,
			expectError: true,
			expectedErr: ErrTokenExpired,
		},
		{
			name:        "Not a refresh token",
			tokenString: accessToken,
			expectError: true,
			expectedErr: ErrNotRefreshToken,
		},
		{
			name:        "Invalid token format",
			tokenString: "invalid.token.format",
			expectError: true,
			expectedErr: ErrInvalidTokenFormat,
		},
		{
			name:        "Empty token",
			tokenString: "",
			expectError: true,
			expectedErr: ErrInvalidTokenFormat,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := ParseRefreshToken(tc.tokenString)

			if tc.expectError {
				assert.Error(t, err)
				var jwtErr *JWTError
				assert.True(t, errors.As(err, &jwtErr), "Expected JWTError type")
				assert.True(t, errors.Is(jwtErr, tc.expectedErr),
					"Expected %v, got %v", tc.expectedErr, jwtErr)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, testUserUUID, claims.UserUUID)
				assert.Equal(t, RefreshToken, claims.TokenType)
				assert.Equal(t, testSessionUUID, claims.SessionUUID)
			}
		})
	}
}

func TestParseToken(t *testing.T) {
	cleanup := setupTestEnv()
	defer cleanup()

	// Create valid access token
	accessClaims := &AccessClaims{
		BaseClaims: BaseClaims{
			UserUUID:           testUserUUID,
			TokenType:          AccessToken,
			ExpirationDuration: time.Hour,
		},
		Permissions: []string{"read"},
	}
	accessToken, err := CreateAccessToken(accessClaims, testSecretKey)
	require.NoError(t, err)

	// Create valid refresh token
	refreshClaims := &RefreshClaims{
		BaseClaims: BaseClaims{
			UserUUID:           testUserUUID,
			TokenType:          RefreshToken,
			ExpirationDuration: time.Hour,
		},
		SessionUUID: testSessionUUID,
	}
	refreshToken, err := CreateRefreshToken(refreshClaims, testSecretKey)
	require.NoError(t, err)

	// Create an expired token
	expiredClaims := &AccessClaims{
		BaseClaims: BaseClaims{
			UserUUID:           testUserUUID,
			TokenType:          AccessToken,
			ExpirationDuration: -time.Hour, // Expired 1 hour ago
		},
		Permissions: []string{"read"},
	}
	expiredToken, err := CreateAccessToken(expiredClaims, testSecretKey)
	require.NoError(t, err)

	tests := []struct {
		name         string
		tokenString  string
		expectError  bool
		expectedErr  error
		expectedType TokenType
	}{
		{
			name:         "Valid access token",
			tokenString:  accessToken,
			expectError:  false,
			expectedType: AccessToken,
		},
		{
			name:         "Valid refresh token",
			tokenString:  refreshToken,
			expectError:  false,
			expectedType: RefreshToken,
		},
		{
			name:        "Expired token",
			tokenString: expiredToken,
			expectError: true,
			expectedErr: ErrTokenExpired,
		},
		{
			name:        "Invalid token format",
			tokenString: "invalid.token.format",
			expectError: true,
			expectedErr: ErrInvalidTokenFormat,
		},
		{
			name:        "Empty token",
			tokenString: "",
			expectError: true,
			expectedErr: ErrInvalidTokenFormat,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := ParseToken(tc.tokenString)

			if tc.expectError {
				assert.Error(t, err)
				var jwtErr *JWTError
				assert.True(t, errors.As(err, &jwtErr), "Expected JWTError type")
				assert.True(t, errors.Is(jwtErr, tc.expectedErr),
					"Expected %v, got %v", tc.expectedErr, jwtErr)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, testUserUUID, claims.UserUUID)
				assert.Equal(t, tc.expectedType, claims.TokenType)
			}
		})
	}
}

func TestJWTErrorMethods(t *testing.T) {
	originalErr := ErrInvalidToken
	message := "test error message"

	jwtErr := &JWTError{
		Err:     originalErr,
		Message: message,
	}

	// Test Error method
	assert.Equal(t, message+": "+originalErr.Error(), jwtErr.Error())

	// Test with empty message
	jwtErrNoMsg := &JWTError{
		Err: originalErr,
	}
	assert.Equal(t, originalErr.Error(), jwtErrNoMsg.Error())

	// Test Unwrap method
	assert.Equal(t, originalErr, jwtErr.Unwrap())

	// Test Is method
	assert.True(t, jwtErr.Is(originalErr))
	assert.False(t, jwtErr.Is(ErrTokenExpired))

	// Test with errors.Is
	assert.True(t, errors.Is(jwtErr, originalErr))
}

func TestKeyFunc(t *testing.T) {
	cleanup := setupTestEnv()
	defer cleanup()

	// Create a token with the correct signing method
	token := jwt.New(jwt.SigningMethodHS512)

	key, err := keyFunc(token)
	assert.NoError(t, err)
	assert.Equal(t, []byte(testSecretKey), key)

	// Test with missing secret
	err = os.Unsetenv("JWT_SECRET")
	assert.NoError(t, err)

	key, err = keyFunc(token)
	assert.Error(t, err)
	assert.Equal(t, ErrMissingSecretKey, err)
	assert.Nil(t, key)

	// Reset the environment
	err = os.Setenv("JWT_SECRET", testSecretKey)
	assert.NoError(t, err)

	// Test with invalid signing method
	invalidToken := jwt.New(jwt.SigningMethodRS256)
	key, err = keyFunc(invalidToken)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidSigningAlg))
	assert.Nil(t, key)
}
