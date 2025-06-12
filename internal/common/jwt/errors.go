package jwt

import (
	"errors"
	"fmt"
)

// Package-wide base error
var ErrJWT = errors.New("jwt error")

// JWT error types
var (
	// General JWT errors
	ErrInvalidToken       = fmt.Errorf("%w: invalid token", ErrJWT)
	ErrTokenExpired       = fmt.Errorf("%w: token has expired", ErrJWT)
	ErrInvalidClaimsType  = fmt.Errorf("%w: invalid claims format", ErrJWT)
	ErrMissingTokenType   = fmt.Errorf("%w: token type not found", ErrJWT)
	ErrMissingSecretKey   = fmt.Errorf("%w: JWT_SECRET environment variable not set", ErrJWT)
	ErrInvalidSigningAlg  = fmt.Errorf("%w: unexpected signing method", ErrJWT)
	ErrInvalidTokenFormat = fmt.Errorf("%w: token format is invalid", ErrJWT)

	// Token type errors
	ErrNotAccessToken  = fmt.Errorf("%w: token is not an access token", ErrJWT)
	ErrNotRefreshToken = fmt.Errorf("%w: token is not a refresh token", ErrJWT)

	// Fields errors
	ErrMissingSessionUUID = fmt.Errorf("%w: session_uuid is missing from token", ErrJWT)
	ErrMissingUserUUID    = fmt.Errorf("%w: user_uuid is missing from token", ErrJWT)

	// Claims validation errors
	ErrInvalidIssuer     = fmt.Errorf("%w: invalid token issuer", ErrJWT)
	ErrInvalidAudience   = fmt.Errorf("%w: invalid token audience", ErrJWT)
	ErrTokenNotValidYet  = fmt.Errorf("%w: token not valid yet", ErrJWT)
	ErrMissingIssuedAt   = fmt.Errorf("%w: issued at claim is missing", ErrJWT)
	ErrMissingExpiration = fmt.Errorf("%w: expiration claim is missing", ErrJWT)
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

func (e *JWTError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// WrapError wraps an error with optional context message
func WrapError(err error, message string) error {
	return &JWTError{
		Err:     err,
		Message: message,
	}
}
