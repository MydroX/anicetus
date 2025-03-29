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
	UserUUID           string
	TokenType          TokenType
	ExpirationDuration time.Duration
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

// CreateAccessToken creates a new access token (without session info)
func CreateAccessToken(c *AccessClaims, secretKey string) (string, error) {
	if secretKey == "" {
		return "", &JWTError{Err: ErrMissingSecretKey}
	}

	expirationTime := time.Now().Add(time.Second * c.ExpirationDuration)
	expT := jwt.NewNumericDate(expirationTime)

	claims := jwt.MapClaims{
		"user_uuid":   c.UserUUID,
		"token_type":  string(AccessToken),
		"permissions": c.Permissions,
		"exp":         expT,
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

	expirationTime := time.Now().Add(time.Second * c.ExpirationDuration)
	expT := jwt.NewNumericDate(expirationTime)

	claims := jwt.MapClaims{
		"user_uuid":    c.UserUUID,
		"session_uuid": c.SessionUUID,
		"token_type":   string(RefreshToken),
		"exp":          expT,
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

// ParseAccessToken parses and validates an access token
func ParseAccessToken(tokenString string) (*AccessClaims, error) {
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
	if perms, ok := claims["permissions"].([]interface{}); ok {
		for _, p := range perms {
			if perm, ok := p.(string); ok {
				permissions = append(permissions, perm)
			}
		}
	}

	// Create and return the access claims
	accessClaims := &AccessClaims{
		BaseClaims: BaseClaims{
			UserUUID:  userUUID,
			TokenType: AccessToken,
		},
		Permissions: permissions,
	}

	return accessClaims, nil
}

// ParseRefreshToken parses and validates a refresh token
func ParseRefreshToken(tokenString string) (*RefreshClaims, error) {
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

	// Create and return the refresh claims
	refreshClaims := &RefreshClaims{
		BaseClaims: BaseClaims{
			UserUUID:  userUUID,
			TokenType: RefreshToken,
		},
		SessionUUID: sessionUUID,
	}

	return refreshClaims, nil
}

// ParseToken parses the token string and returns the base claims
func ParseToken(tokenString string) (*BaseClaims, error) {
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

	// Extract token type with validation
	tokenType, ok := claims["token_type"].(string)
	if !ok {
		return nil, &JWTError{Err: ErrMissingTokenType}
	}

	userUUID, ok := claims["user_uuid"].(string)
	if !ok || userUUID == "" {
		return nil, &JWTError{Err: ErrMissingUserUUID}
	}

	// Create and return the base claims
	baseClaims := &BaseClaims{
		UserUUID:  userUUID,
		TokenType: TokenType(tokenType),
	}

	return baseClaims, nil
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
