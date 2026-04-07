package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_DEV(t *testing.T) {
	l, err := New("DEV")
	require.NoError(t, err)
	assert.NotNil(t, l)
}

func TestNew_PROD(t *testing.T) {
	l, err := New("PROD")
	require.NoError(t, err)
	assert.NotNil(t, l)
}

func TestNew_TEST(t *testing.T) {
	l, err := New("TEST")
	require.NoError(t, err)
	assert.NotNil(t, l)
}

func TestNew_InvalidEnv(t *testing.T) {
	l, err := New("INVALID")
	assert.Error(t, err)
	assert.Nil(t, l)
	assert.Contains(t, err.Error(), "invalid environment")
}

func TestNew_EmptyEnv(t *testing.T) {
	l, err := New("")
	assert.Error(t, err)
	assert.Nil(t, l)
}
