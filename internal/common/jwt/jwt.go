package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

//go:generate mockgen -source=jwt.go -destination=mock_jwt.go -package=jwt

// TokenService defines the interface for JWT token operations
type TokenService interface {
	// CreateAccessToken creates a new access token
	CreateAccessToken(userUUID string, permissions []string, expiration time.Duration) (string, error)

	// CreateRefreshToken creates a refresh token with session info
	CreateRefreshToken(userUUID string, sessionUUID string, expiration time.Duration) (string, error)

	// ParseAccessToken parses and validates an access token
	ParseAccessToken(tokenString string) (*AccessClaims, error)

	// ParseRefreshToken parses and validates a refresh token
	ParseRefreshToken(tokenString string) (*RefreshClaims, error)

	// ParseToken parses a token and returns base claims without validating token type
	ParseToken(tokenString string) (*BaseClaims, error)
}

// AudienceProvider defines the interface for fetching allowed audiences
type AudienceProvider interface {
	// GetAllowedAudiences returns a list of allowed JWT audience values
	GetAllowedAudiences() ([]string, error)
}

// TokenType defines the type of JWT token
type TokenType string

// Defined token types
const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type BaseClaims struct {
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
	Permissions []string
}

// RefreshClaims for the long-lived refresh token that maintains session
type RefreshClaims struct {
	BaseClaims
	SessionUUID string
}

// AccessTokenClaims combines registered claims with custom claims for access tokens
type AccessTokenClaims struct {
	jwt.RegisteredClaims
	UserUUID    string   `json:"user_uuid"`
	TokenType   string   `json:"token_type"`
	Permissions []string `json:"permissions"`
}

// RefreshTokenClaims combines registered claims with custom claims for refresh tokens
type RefreshTokenClaims struct {
	jwt.RegisteredClaims
	UserUUID    string `json:"user_uuid"`
	TokenType   string `json:"token_type"`
	SessionUUID string `json:"session_uuid"`
}
