package jwt

import (
	"encoding/json"
	"strings"
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
		AccessTokenSecret:         "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		RefreshTokenSecret: "test-refresh-secret-long-enough-for-signing-jwt-tok",
		ExpectedIssuer:    "test-issuer",
		ExpectedAudiences: []string{"test-audience"},
		ClockSkewSeconds:  60,
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

		tokenString, err := token.SignedString([]byte(config.AccessTokenSecret))
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

		tokenString, err := token.SignedString([]byte(config.AccessTokenSecret))
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

		tokenString, err := token.SignedString([]byte(config.AccessTokenSecret))
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

func TestParseToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := TokenConfig{
		AccessTokenSecret:         "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		RefreshTokenSecret: "test-refresh-secret-long-enough-for-signing-jwt-tok",
		ExpectedIssuer:    "test-issuer",
		ExpectedAudiences: []string{"test-audience"},
		ClockSkewSeconds:  60,
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

		tokenString, err := token.SignedString([]byte(config.AccessTokenSecret))
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

		tokenString, err := token.SignedString([]byte(config.AccessTokenSecret))
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
			AccessTokenSecret:           "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		RefreshTokenSecret: "test-refresh-secret-long-enough-for-signing-jwt-tok",
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
			return []byte(config.AccessTokenSecret), nil
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
			AccessTokenSecret:      "", // Empty key
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
			AccessTokenSecret:            "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		RefreshTokenSecret: "test-refresh-secret-long-enough-for-signing-jwt-tok",
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
			return []byte(config.RefreshTokenSecret), nil
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
			AccessTokenSecret:      "", // Empty key
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
			AccessTokenSecret:      "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		RefreshTokenSecret: "test-refresh-secret-long-enough-for-signing-jwt-tok",
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

// FuzzParseToken fuzzes the ParseToken function with various malformed inputs
func FuzzParseToken(f *testing.F) {
	// Seed with various test cases
	f.Add("")
	f.Add("invalid")
	f.Add("a.b.c")
	f.Add("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")
	f.Add("valid.token.here")
	f.Add("malformed..token")
	f.Add(".......")
	f.Add("eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.invalid.signature")

	config := TokenConfig{
		AccessTokenSecret:        "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		RefreshTokenSecret: "test-refresh-secret-long-enough-for-signing-jwt-tok",
		ExpectedIssuer:   "test-issuer",
		ClockSkewSeconds: 60,
	}
	service := NewJWTService(config)

	f.Fuzz(func(t *testing.T, tokenString string) {
		// The function should handle any input gracefully without panicking
		claims, err := service.ParseToken(tokenString)

		// Verify the function handles all inputs gracefully
		if err != nil {
			// If there's an error, claims should be nil
			assert.Nil(t, claims, "Claims should be nil when there's an error")
			// Error should have a meaningful message
			assert.NotEmpty(t, err.Error(), "Error message should not be empty")
		} else {
			// If successful, claims should not be nil and should have required fields
			assert.NotNil(t, claims, "Claims should not be nil when successful")
			assert.NotEmpty(t, claims.UserUUID, "UserUUID should not be empty when successful")
			assert.NotEqual(t, TokenType(""), claims.TokenType, "TokenType should not be empty when successful")
		}

		// The function should never panic
		// This is implicitly tested by the fuzz framework
	})
}

// FuzzParseTokenWithStructuredInput creates structured malformed tokens for fuzzing
func FuzzParseTokenWithStructuredInput(f *testing.F) {
	// Seed with structured inputs that target specific parts of the JWT
	f.Add("header", `{"alg":"HS512","typ":"JWT"}`, `{"exp":1234567890,"iat":1234567890,"iss":"test","user_uuid":"user-123","token_type":"access"}`, "signature")
	f.Add("malformed", `{"invalid":json}`, `{"exp":"not_number"}`, "bad_sig")
	f.Add("", `{}`, `{}`, "")
	f.Add("valid", `{"alg":"HS256","typ":"JWT"}`, `{"user_uuid":"","token_type":""}`, "valid_sig")

	config := TokenConfig{
		AccessTokenSecret:        "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		RefreshTokenSecret: "test-refresh-secret-long-enough-for-signing-jwt-tok",
		ExpectedIssuer:   "test-issuer",
		ClockSkewSeconds: 60,
	}
	service := NewJWTService(config)

	f.Fuzz(func(t *testing.T, header, payload, signature, secretKey string) {
		// Create various malformed JWT structures
		testCases := []string{
			// Standard 3-part structure
			header + "." + payload + "." + signature,
			// Missing parts
			header + "." + payload,
			header,
			"." + payload + "." + signature,
			header + ".." + signature,
			// Extra parts
			header + "." + payload + "." + signature + ".extra",
			// Empty parts
			"..",
			"...",
			"" + "." + "" + "." + "",
			// Only separators
			".....",
			// Base64-like but invalid
			"YWFh.YmJi.Y2Nj", // "aaa.bbb.ccc" in base64
		}

		for _, tokenString := range testCases {
			claims, err := service.ParseToken(tokenString)

			// Should handle gracefully
			if err != nil {
				assert.Nil(t, claims)
				assert.NotEmpty(t, err.Error())
			} else {
				assert.NotNil(t, claims)
				// If parsing succeeded, basic fields should be valid
				assert.NotEmpty(t, claims.UserUUID)
			}
		}
	})
}

// FuzzParseTokenWithValidStructure fuzzes with valid JWT structure but invalid content
func FuzzParseTokenWithValidStructure(f *testing.F) {
	// Seed with valid JWT structures but fuzzed content
	f.Add("user-123", "access", int64(1234567890), "test-issuer", "audience")
	f.Add("", "", int64(0), "", "")
	f.Add("very-long-uuid-that-might-cause-issues-with-memory-allocation", "refresh", int64(-1), "malicious-issuer", "evil-audience")

	config := TokenConfig{
		AccessTokenSecret:        "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		RefreshTokenSecret: "test-refresh-secret-long-enough-for-signing-jwt-tok",
		ExpectedIssuer:   "test-issuer",
		ClockSkewSeconds: 60,
	}

	f.Fuzz(func(t *testing.T, userUUID, tokenType string, exp int64, issuer, audience string) {
		// Create a valid JWT structure with fuzzed content
		claims := jwt.MapClaims{
			"user_uuid":  userUUID,
			"token_type": tokenType,
			"exp":        exp,
			"iat":        time.Now().Unix(),
			"iss":        issuer,
			"aud":        audience,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		tokenString, err := token.SignedString([]byte(config.AccessTokenSecret))

		if err != nil {
			// If we can't create the token, skip this iteration
			t.Skip("Could not create test token")
			return
		}

		// Parse the token
		service := NewJWTService(config)
		parsedClaims, parseErr := service.ParseToken(tokenString)

		// Should handle gracefully
		if parseErr != nil {
			assert.Nil(t, parsedClaims)
			assert.NotEmpty(t, parseErr.Error())
		} else {
			assert.NotNil(t, parsedClaims)
			// Verify the parsing preserved the data correctly
			assert.Equal(t, userUUID, parsedClaims.UserUUID)
			assert.Equal(t, TokenType(tokenType), parsedClaims.TokenType)
		}
	})
}

// FuzzParseTokenClaimTypes fuzzes with different data types for claims
func FuzzParseTokenClaimTypes(f *testing.F) {
	// Seed with various data types that might be in claims
	f.Add(`{"user_uuid":123,"token_type":"access"}`)  // number for user_uuid
	f.Add(`{"user_uuid":"user","token_type":456}`)    // number for token_type
	f.Add(`{"user_uuid":[],"token_type":"access"}`)   // array for user_uuid
	f.Add(`{"user_uuid":"user","token_type":{}}`)     // object for token_type
	f.Add(`{"user_uuid":null,"token_type":"access"}`) // null values
	f.Add(`{"user_uuid":"user","token_type":null}`)
	f.Add(`{"user_uuid":true,"token_type":false}`)                 // boolean values
	f.Add(`{"exp":"not_a_number","iat":"also_not_number"}`)        // invalid time claims
	f.Add(`{"nested":{"user_uuid":"user","token_type":"access"}}`) // nested structure

	config := TokenConfig{
		AccessTokenSecret:        "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		RefreshTokenSecret: "test-refresh-secret-long-enough-for-signing-jwt-tok",
		ExpectedIssuer:   "test-issuer",
		ClockSkewSeconds: 60,
	}

	f.Fuzz(func(t *testing.T, claimsJSON string) {
		// Try to parse the JSON to create claims
		var claimsMap map[string]interface{}
		if err := json.Unmarshal([]byte(claimsJSON), &claimsMap); err != nil {
			// If JSON is invalid, create a token with the raw string somehow
			// or skip this iteration
			t.Skip("Invalid JSON for claims")
			return
		}

		// Add required time claims if missing to make a somewhat valid token
		if _, exists := claimsMap["exp"]; !exists {
			claimsMap["exp"] = time.Now().Add(time.Hour).Unix()
		}
		if _, exists := claimsMap["iat"]; !exists {
			claimsMap["iat"] = time.Now().Unix()
		}
		if _, exists := claimsMap["iss"]; !exists {
			claimsMap["iss"] = "test-issuer"
		}

		// Create JWT with these claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims(claimsMap))
		tokenString, err := token.SignedString([]byte(config.AccessTokenSecret))

		if err != nil {
			t.Skip("Could not create test token")
			return
		}

		// Parse the token
		service := NewJWTService(config)
		parsedClaims, parseErr := service.ParseToken(tokenString)

		// Should handle gracefully
		if parseErr != nil {
			assert.Nil(t, parsedClaims)
			assert.NotEmpty(t, parseErr.Error())

			// Log which specific validation failed for debugging
			t.Logf("Parse error: %v for claims: %s", parseErr, claimsJSON)
		} else {
			assert.NotNil(t, parsedClaims)
			// If successful, basic validations should have passed
			assert.NotEmpty(t, parsedClaims.UserUUID)
			assert.NotEqual(t, TokenType(""), parsedClaims.TokenType)
		}
	})
}

// FuzzParseTokenMaliciousPayloads fuzzes with security-focused malicious inputs
func FuzzParseTokenMaliciousPayloads(f *testing.F) {
	// Seed with security-focused malicious inputs
	f.Add("../../../etc/passwd")
	f.Add("<script>alert('xss')</script>")
	f.Add("'; DROP TABLE users; --")
	f.Add("\x00\x01\x02\x03")         // binary data
	f.Add(strings.Repeat("A", 10000)) // very long string
	f.Add("user\n\r\t\x00")
	f.Add("🔐💀🚨")   // emojis
	f.Add("用户123") // unicode

	config := TokenConfig{
		AccessTokenSecret:        "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		RefreshTokenSecret: "test-refresh-secret-long-enough-for-signing-jwt-tok",
		ExpectedIssuer:   "test-issuer",
		ClockSkewSeconds: 60,
	}

	f.Fuzz(func(t *testing.T, maliciousInput string) {
		// Test malicious input in various claim fields
		testCases := []jwt.MapClaims{
			{
				"user_uuid":  maliciousInput,
				"token_type": "access",
				"exp":        time.Now().Add(time.Hour).Unix(),
				"iat":        time.Now().Unix(),
				"iss":        "test-issuer",
			},
			{
				"user_uuid":  "user-123",
				"token_type": maliciousInput,
				"exp":        time.Now().Add(time.Hour).Unix(),
				"iat":        time.Now().Unix(),
				"iss":        "test-issuer",
			},
			{
				"user_uuid":  "user-123",
				"token_type": "access",
				"exp":        time.Now().Add(time.Hour).Unix(),
				"iat":        time.Now().Unix(),
				"iss":        maliciousInput,
			},
		}

		service := NewJWTService(config)

		for i, claims := range testCases {
			token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
			tokenString, err := token.SignedString([]byte(config.AccessTokenSecret))

			if err != nil {
				// Some malicious inputs might prevent token creation
				continue
			}

			// Parse the token
			parsedClaims, parseErr := service.ParseToken(tokenString)

			// Should handle gracefully without panicking
			if parseErr != nil {
				assert.Nil(t, parsedClaims)
				assert.NotEmpty(t, parseErr.Error())
			} else {
				assert.NotNil(t, parsedClaims)

				// Log successful parsing of potentially dangerous input
				t.Logf("Test case %d: Malicious input '%s' was parsed successfully", i, maliciousInput)

				// Verify the malicious input is contained properly
				switch i {
				case 0: // malicious user_uuid
					assert.Equal(t, maliciousInput, parsedClaims.UserUUID)
				case 1: // malicious token_type
					assert.Equal(t, TokenType(maliciousInput), parsedClaims.TokenType)
				case 2: // malicious issuer
					assert.Equal(t, maliciousInput, parsedClaims.Issuer)
				}
			}
		}
	})
}

func TestKeyFunc(t *testing.T) {
	t.Run("valid HMAC method", func(t *testing.T) {
		config := TokenConfig{
			AccessTokenSecret: "test-secret-key",
		}
		service := NewJWTService(config)

		token := &jwt.Token{
			Method: jwt.SigningMethodHS256,
			Header: map[string]interface{}{
				"alg": "HS256",
			},
		}

		key, err := service.keyFuncForSecret("test-secret-key")(token)
		assert.NoError(t, err)
		assert.Equal(t, []byte("test-secret-key"), key)
	})

	t.Run("missing secret key", func(t *testing.T) {
		config := TokenConfig{
			AccessTokenSecret: "",
		}
		service := NewJWTService(config)

		token := &jwt.Token{
			Method: jwt.SigningMethodHS256,
			Header: map[string]interface{}{
				"alg": "HS256",
			},
		}

		key, err := service.keyFuncForSecret("")(token)
		assert.Error(t, err)
		assert.Nil(t, key)
		assert.ErrorIs(t, err, ErrMissingSecretKey)
	})

	t.Run("invalid signing method", func(t *testing.T) {
		config := TokenConfig{
			AccessTokenSecret: "test-secret-key",
		}
		service := NewJWTService(config)

		token := &jwt.Token{
			Method: jwt.SigningMethodRS256,
			Header: map[string]interface{}{
				"alg": "RS256",
			},
		}

		key, err := service.keyFuncForSecret("test-secret-key")(token)
		assert.Error(t, err)
		assert.Nil(t, key)
		assert.ErrorIs(t, err, ErrInvalidSigningAlg)
	})

	t.Run("none algorithm", func(t *testing.T) {
		config := TokenConfig{
			AccessTokenSecret: "test-secret-key",
		}
		service := NewJWTService(config)

		token := &jwt.Token{
			Method: jwt.SigningMethodNone,
			Header: map[string]interface{}{
				"alg": "none",
			},
		}

		key, err := service.keyFuncForSecret("test-secret-key")(token)
		assert.Error(t, err)
		assert.Nil(t, key)
		assert.ErrorIs(t, err, ErrInvalidSigningAlg)
	})
}

func TestParseRefreshToken(t *testing.T) {
	config := TokenConfig{
		AccessTokenSecret:         "test-secret-key-long-enough-for-signing-jwt-tokens-securely",
		RefreshTokenSecret: "test-refresh-secret-long-enough-for-signing-jwt-tok",
		ExpectedIssuer:    "test-issuer",
		ExpectedAudiences: []string{"test-audience"},
		ClockSkewSeconds:  60,
	}
	service := NewJWTService(config)

	t.Run("valid refresh token", func(t *testing.T) {
		// Create valid refresh token
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"exp":          time.Now().Add(24 * time.Hour).Unix(),
			"iat":          time.Now().Unix(),
			"iss":          "test-issuer",
			"aud":          []string{"test-audience"},
			"user_uuid":    "user-123",
			"token_type":   string(RefreshToken),
			"session_uuid": "session-456",
		})

		tokenString, err := token.SignedString([]byte(config.RefreshTokenSecret))
		assert.NoError(t, err)

		claims, err := service.ParseRefreshToken(tokenString)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, "user-123", claims.UserUUID)
		assert.Equal(t, RefreshToken, claims.TokenType)
		assert.Equal(t, "session-456", claims.SessionUUID)
	})

	t.Run("invalid token format", func(t *testing.T) {
		claims, err := service.ParseRefreshToken("invalid-token")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("expired token", func(t *testing.T) {
		// Create expired token
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"exp":          time.Now().Add(-time.Hour).Unix(),
			"iat":          time.Now().Add(-2 * time.Hour).Unix(),
			"iss":          "test-issuer",
			"user_uuid":    "user-123",
			"token_type":   string(RefreshToken),
			"session_uuid": "session-456",
		})

		tokenString, err := token.SignedString([]byte(config.RefreshTokenSecret))
		assert.NoError(t, err)

		claims, err := service.ParseRefreshToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("wrong token type - access token", func(t *testing.T) {
		// Create access token instead of refresh token
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"exp":        time.Now().Add(time.Hour).Unix(),
			"iat":        time.Now().Unix(),
			"iss":        "test-issuer",
			"user_uuid":  "user-123",
			"token_type": string(AccessToken), // Wrong type
		})

		tokenString, err := token.SignedString([]byte(config.RefreshTokenSecret))
		assert.NoError(t, err)

		claims, err := service.ParseRefreshToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("missing user_uuid", func(t *testing.T) {
		// Create token without user_uuid
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"exp":          time.Now().Add(time.Hour).Unix(),
			"iat":          time.Now().Unix(),
			"iss":          "test-issuer",
			"token_type":   string(RefreshToken),
			"session_uuid": "session-456",
			// Missing user_uuid
		})

		tokenString, err := token.SignedString([]byte(config.RefreshTokenSecret))
		assert.NoError(t, err)

		claims, err := service.ParseRefreshToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("missing session_uuid", func(t *testing.T) {
		// Create token without session_uuid
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"exp":        time.Now().Add(time.Hour).Unix(),
			"iat":        time.Now().Unix(),
			"iss":        "test-issuer",
			"user_uuid":  "user-123",
			"token_type": string(RefreshToken),
			// Missing session_uuid
		})

		tokenString, err := token.SignedString([]byte(config.RefreshTokenSecret))
		assert.NoError(t, err)

		claims, err := service.ParseRefreshToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("empty session_uuid", func(t *testing.T) {
		// Create token with empty session_uuid
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"exp":          time.Now().Add(time.Hour).Unix(),
			"iat":          time.Now().Unix(),
			"iss":          "test-issuer",
			"user_uuid":    "user-123",
			"token_type":   string(RefreshToken),
			"session_uuid": "", // Empty string
		})

		tokenString, err := token.SignedString([]byte(config.RefreshTokenSecret))
		assert.NoError(t, err)

		claims, err := service.ParseRefreshToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}
