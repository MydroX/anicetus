package hmac

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashWithSalt(t *testing.T) {
	salt := base64.StdEncoding.EncodeToString([]byte("test-secret-key!"))

	hash, err := HashWithSalt("hello", salt)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Verify output is valid base64
	_, err = base64.StdEncoding.DecodeString(hash)
	assert.NoError(t, err)
}

func TestHashWithSalt_Consistency(t *testing.T) {
	salt := base64.StdEncoding.EncodeToString([]byte("consistent-key!!"))

	hash1, err := HashWithSalt("same-input", salt)
	require.NoError(t, err)

	hash2, err := HashWithSalt("same-input", salt)
	require.NoError(t, err)

	assert.Equal(t, hash1, hash2)
}

func TestHashWithSalt_DifferentInputs(t *testing.T) {
	salt := base64.StdEncoding.EncodeToString([]byte("shared-key-value"))

	hash1, err := HashWithSalt("input-a", salt)
	require.NoError(t, err)

	hash2, err := HashWithSalt("input-b", salt)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestHashWithSalt_InvalidBase64(t *testing.T) {
	_, err := HashWithSalt("hello", "not-valid-base64!!!")
	assert.Error(t, err)
}

func TestHashWithSalt_EmptyInput(t *testing.T) {
	salt := base64.StdEncoding.EncodeToString([]byte("some-key-value!!"))

	hash, err := HashWithSalt("", salt)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}
