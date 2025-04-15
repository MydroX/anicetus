//revive:disable:cognitive-complexity
//revive:disable:add-constant
package jwt

import (
	"MydroX/anicetus/pkg/argon2id"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenJWT string
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claim keys
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

// Custom error types for better error handling
var (
	// General JWT errors
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token has expired")
	ErrInvalidClaimsType  = errors.New("invalid claims format")
	ErrMissingTokenType   = errors.New("token type not found")
	ErrMissingSecretKey   = errors.New("JWT_SECRET environment variable not set")
	ErrInvalidSigningAlg  = errors.New("unexpected signing method")
	ErrInvalidTokenFormat = errors.New("token format is invalid")

	// Token type errors
	ErrNotAccessToken  = errors.New("token is not an access token")
	ErrNotRefreshToken = errors.New("token is not a refresh token")

	// Fields errors
	ErrMissingSessionUUID = errors.New("session_uuid is missing from token")
	ErrMissingUserUUID    = errors.New("user_uuid is missing from token")

	// Claims validation errors
	ErrInvalidIssuer     = errors.New("invalid token issuer")
	ErrInvalidAudience   = errors.New("invalid token audience")
	ErrTokenNotValidYet  = errors.New("token not valid yet")
	ErrMissingIssuedAt   = errors.New("issued at claim is missing")
	ErrMissingExpiration = errors.New("expiration claim is missing")
)

// JWTError provides more context about JWT errors
type JWTError struct {
	Err     error
	Message string
}

func (e *JWTError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Err.Error()
}

func (e *JWTError) Unwrap() error {
	return e.Err
}

func (e *JWTError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// BaseClaims contains essential fields common to all tokens
type BaseClaims struct {
	JWTID     string
	UserUUID  string
	TokenType TokenType
	Exp       int64
	IssuedAt  time.Time
	Issuer    string
	Audience  string
}

// AccessClaims for the short-lived access token
type AccessClaims struct {
	BaseClaims
	Role        []string
	Permissions []string
}

// RefreshClaims for the long-lived refresh token that maintains session
type RefreshClaims struct {
	BaseClaims
	SessionUUID string
}

// TokenConfig holds configuration for token validation
type TokenConfig struct {
	ExpectedIssuer   string
	ExpectedAudience string
	ClockSkewSeconds int // Tolerance for time-based claims
}

// CreateAccessToken creates a new access token (without session info)
func CreateAccessToken(c *AccessClaims, secretKey string) (string, error) {
	if secretKey == "" {
		return "", &JWTError{Err: ErrMissingSecretKey}
	}

	expirationTime := time.Now().Add(time.Duration(c.Exp) * time.Second)
	expT := jwt.NewNumericDate(expirationTime)

	claims := jwt.MapClaims{
		"user_uuid":   c.UserUUID,
		"token_type":  string(AccessToken),
		"permissions": c.Permissions,
		"exp":         expT,
		"iss":         c.Issuer,
		"aud":         c.Audience,
		"iat":         jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	ss, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", &JWTError{
			Err:     err,
			Message: "failed to sign access token",
		}
	}

	return ss, nil
}

// CreateRefreshToken creates a new refresh token with session info
func CreateRefreshToken(c *RefreshClaims, secretKey string) (string, error) {
	if secretKey == "" {
		return "", &JWTError{Err: ErrMissingSecretKey}
	}

	expirationTime := time.Now().Add(time.Duration(c.Exp) * time.Second)
	expT := jwt.NewNumericDate(expirationTime)

	claims := jwt.MapClaims{
		"user_uuid":    c.UserUUID,
		"session_uuid": c.SessionUUID,
		"token_type":   string(RefreshToken),
		"exp":          expT,
		"iss":          c.Issuer,
		"aud":          c.Audience,
		"iat":          jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	ss, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", &JWTError{
			Err:     err,
			Message: "failed to sign refresh token",
		}
	}

	return ss, nil
}

// validateStandardClaims validates standard JWT claims
func validateStandardClaims(claims jwt.MapClaims, config *TokenConfig) error {
	now := time.Now()
	clockSkew := time.Duration(config.ClockSkewSeconds) * time.Second

	// Validate issuer if present
	if iss, ok := claims["iss"].(string); ok {
		if iss != config.ExpectedIssuer {
			return &JWTError{
				Err:     ErrInvalidIssuer,
				Message: fmt.Sprintf("expected %s, got %s", config.ExpectedIssuer, iss),
			}
		}
	}

	// Validate audience if present
	if aud, ok := claims["aud"].(string); ok {
		if aud != config.ExpectedAudience {
			return &JWTError{
				Err:     ErrInvalidAudience,
				Message: fmt.Sprintf("expected %s, got %s", config.ExpectedAudience, aud),
			}
		}
	} else if audArray, ok := claims["aud"].([]any); ok {
		validAudience := false
		for _, a := range audArray {
			if aud, ok := a.(string); ok && aud == config.ExpectedAudience {
				validAudience = true
				break
			}
		}
		if !validAudience {
			return &JWTError{Err: ErrInvalidAudience}
		}
	}

	// Validate expiration time (required)
	if exp, ok := claims["exp"].(float64); !ok {
		return &JWTError{Err: ErrMissingExpiration}
	} else {
		expTime := time.Unix(int64(exp), 0)
		if now.After(expTime.Add(clockSkew)) {
			return &JWTError{Err: ErrTokenExpired}
		}
	}

	// Validate issued at if present
	if iat, ok := claims["iat"].(float64); ok {
		issuedAt := time.Unix(int64(iat), 0)
		if now.Before(issuedAt.Add(-clockSkew)) {
			return &JWTError{
				Err:     ErrMissingIssuedAt,
				Message: "token used before issued",
			}
		}
	}

	// Validate not before if present
	if nbf, ok := claims["nbf"].(float64); ok {
		notBefore := time.Unix(int64(nbf), 0)
		if now.Before(notBefore.Add(-clockSkew)) {
			return &JWTError{Err: ErrTokenNotValidYet}
		}
	}

	return nil
}

// ParseAccessToken parses and validates an access token
func ParseAccessToken(tokenString string, config *TokenConfig) (*AccessClaims, error) {
	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, &JWTError{Err: ErrTokenExpired}
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, &JWTError{Err: ErrInvalidTokenFormat}
		}
		return nil, &JWTError{Err: err, Message: "failed to parse access token"}
	}

	if !token.Valid {
		return nil, &JWTError{Err: ErrInvalidToken}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, &JWTError{Err: ErrInvalidClaimsType}
	}

	// Validate standard claims
	if err := validateStandardClaims(claims, config); err != nil {
		return nil, err
	}

	// Verify this is an access token
	tokenType, ok := claims["token_type"].(string)
	if !ok {
		return nil, &JWTError{Err: ErrMissingTokenType}
	}

	if TokenType(tokenType) != AccessToken {
		return nil, &JWTError{Err: ErrNotAccessToken}
	}

	userUUID, ok := claims["user_uuid"].(string)
	if !ok || userUUID == "" {
		return nil, &JWTError{Err: ErrMissingUserUUID}
	}

	// Extract permissions
	var permissions []string
	if perms, ok := claims["permissions"].([]any); ok {
		for _, p := range perms {
			if perm, ok := p.(string); ok {
				permissions = append(permissions, perm)
			}
		}
	}

	// Extract standard claims for BaseClaims
	var issuedAt time.Time
	if iat, ok := claims["iat"].(float64); ok {
		issuedAt = time.Unix(int64(iat), 0)
	}

	var expiration int64
	if exp, ok := claims["exp"].(float64); ok {
		expiration = int64(exp)
	}

	issuer, _ := claims["iss"].(string)
	audience, _ := claims["aud"].(string)

	// Create and return the access claims
	accessClaims := &AccessClaims{
		BaseClaims: BaseClaims{
			UserUUID:  userUUID,
			TokenType: AccessToken,
			Exp:       expiration,
			IssuedAt:  issuedAt,
			Issuer:    issuer,
			Audience:  audience,
		},
		Permissions: permissions,
	}

	return accessClaims, nil
}

// ParseRefreshToken parses and validates a refresh token
func ParseRefreshToken(tokenString string, config *TokenConfig) (*RefreshClaims, error) {
	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, &JWTError{Err: ErrTokenExpired}
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, &JWTError{Err: ErrInvalidTokenFormat}
		}
		return nil, &JWTError{Err: err, Message: "failed to parse refresh token"}
	}

	if !token.Valid {
		return nil, &JWTError{Err: ErrInvalidToken}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, &JWTError{Err: ErrInvalidClaimsType}
	}

	// Validate standard claims
	if err := validateStandardClaims(claims, config); err != nil {
		return nil, err
	}

	// Verify this is a refresh token
	tokenType, ok := claims["token_type"].(string)
	if !ok {
		return nil, &JWTError{Err: ErrMissingTokenType}
	}

	if TokenType(tokenType) != RefreshToken {
		return nil, &JWTError{Err: ErrNotRefreshToken}
	}

	// Extract required fields with validation
	sessionUUID, ok := claims["session_uuid"].(string)
	if !ok || sessionUUID == "" {
		return nil, &JWTError{Err: ErrMissingSessionUUID}
	}

	userUUID, ok := claims["user_uuid"].(string)
	if !ok || userUUID == "" {
		return nil, &JWTError{Err: ErrMissingUserUUID}
	}

	// Extract standard claims for BaseClaims
	var issuedAt time.Time
	if iat, ok := claims["iat"].(float64); ok {
		issuedAt = time.Unix(int64(iat), 0)
	}

	var expiration int64
	if exp, ok := claims["exp"].(float64); ok {
		expiration = int64(exp)
	}

	issuer, _ := claims["iss"].(string)
	audience, _ := claims["aud"].(string)

	// Create and return the refresh claims
	refreshClaims := &RefreshClaims{
		BaseClaims: BaseClaims{
			UserUUID:  userUUID,
			TokenType: RefreshToken,
			Exp:       expiration,
			IssuedAt:  issuedAt,
			Issuer:    issuer,
			Audience:  audience,
		},
		SessionUUID: sessionUUID,
	}

	return refreshClaims, nil
}

// ParseToken parses the token string and returns the base claims
func ParseToken(tokenString string, config *TokenConfig) (*BaseClaims, error) {
	claims, err := parseAndValidateToken(tokenString, config)
	if err != nil {
		return nil, err
	}

	// Extract token type with validation
	tokenType, ok := claims[claimTokenType].(string)
	if !ok {
		return nil, &JWTError{Err: ErrMissingTokenType}
	}

	userUUID, ok := claims[claimUserUUID].(string)
	if !ok || userUUID == "" {
		return nil, &JWTError{Err: ErrMissingUserUUID}
	}

	// Create base claims
	baseClaims := &BaseClaims{
		UserUUID:  userUUID,
		TokenType: TokenType(tokenType),
		Exp:       extractInt64Claim(claims, claimExp),
		IssuedAt:  extractTimeClaim(claims, claimIAT),
		Issuer:    extractStringClaim(claims, claimIss),
		Audience:  extractStringClaim(claims, claimAud),
	}

	return baseClaims, nil
}

// Helper functions for extracting claims
func extractStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}

func extractInt64Claim(claims jwt.MapClaims, key string) int64 {
	if val, ok := claims[key].(float64); ok {
		return int64(val)
	}
	return 0
}

func extractTimeClaim(claims jwt.MapClaims, key string) time.Time {
	if val, ok := claims[key].(float64); ok {
		return time.Unix(int64(val), 0)
	}
	return time.Time{}
}

// Common function for parsing and validating tokens
func parseAndValidateToken(tokenString string, config *TokenConfig) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, &JWTError{Err: ErrTokenExpired}
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, &JWTError{Err: ErrInvalidTokenFormat}
		}
		return nil, &JWTError{Err: err, Message: "failed to parse token"}
	}

	if !token.Valid {
		return nil, &JWTError{Err: ErrInvalidToken}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, &JWTError{Err: ErrInvalidClaimsType}
	}

	// Validate standard claims
	if err := validateStandardClaims(claims, config); err != nil {
		return nil, err
	}

	return claims, nil
}

func (t TokenJWT) String() string {
	return string(t)
}

// HashArgon2 is a function to hash the token with a argon2 algorithm.
func (t TokenJWT) HashArgon2(params *argon2id.HashParams) (string, error) {
	return argon2id.Hash(t.String(), params)
}

func keyFunc(token *jwt.Token) (any, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return nil, ErrMissingSecretKey
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSigningAlg, token.Header["alg"])
	}

	return []byte(secretKey), nil
}
