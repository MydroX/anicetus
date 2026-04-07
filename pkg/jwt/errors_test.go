package jwt

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJWTError_Error_WithMessage(t *testing.T) {
	err := &JWTError{
		Err:     ErrInvalidToken,
		Message: "token verification failed",
	}
	assert.Equal(t, "token verification failed: jwt error: invalid token", err.Error())
}

func TestJWTError_Error_WithoutMessage(t *testing.T) {
	err := &JWTError{
		Err:     ErrInvalidToken,
		Message: "",
	}
	assert.Equal(t, ErrInvalidToken.Error(), err.Error())
}

func TestJWTError_Is(t *testing.T) {
	err := &JWTError{Err: ErrJWT}
	assert.True(t, errors.Is(err, ErrJWT))
}

func TestJWTError_Is_Nested(t *testing.T) {
	err := &JWTError{Err: ErrInvalidToken}
	assert.True(t, errors.Is(err, ErrJWT))
	assert.True(t, errors.Is(err, ErrInvalidToken))
}

func TestJWTError_Is_NoMatch(t *testing.T) {
	err := &JWTError{Err: ErrInvalidToken}
	assert.False(t, errors.Is(err, errors.New("unrelated")))
}

func TestWrapError(t *testing.T) {
	err := WrapError(ErrTokenExpired, "refresh failed")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTokenExpired))
	assert.Contains(t, err.Error(), "refresh failed")
}

func TestWrapError_EmptyMessage(t *testing.T) {
	err := WrapError(ErrInvalidToken, "")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidToken))
	assert.Equal(t, ErrInvalidToken.Error(), err.Error())
}

func TestSentinelErrors(t *testing.T) {
	sentinels := []error{
		ErrInvalidToken,
		ErrTokenExpired,
		ErrInvalidClaimsType,
		ErrMissingTokenType,
		ErrMissingSecretKey,
		ErrInvalidSigningAlg,
		ErrInvalidTokenFormat,
		ErrNotAccessToken,
		ErrNotRefreshToken,
		ErrMissingSessionUUID,
		ErrMissingUserUUID,
		ErrInvalidIssuer,
		ErrInvalidAudience,
		ErrTokenNotValidYet,
		ErrMissingIssuedAt,
		ErrMissingExpiration,
	}

	for _, sentinel := range sentinels {
		t.Run(sentinel.Error(), func(t *testing.T) {
			assert.True(t, errors.Is(sentinel, ErrJWT), "sentinel %v should wrap ErrJWT", sentinel)
		})
	}
}

func TestHandleParseError_Nil(t *testing.T) {
	err := handleParseError(nil, "access token")
	assert.Nil(t, err)
}
