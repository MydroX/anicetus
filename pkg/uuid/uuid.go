package uuid

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// NewWithPrefix generates a new UUID string with a prefix
func NewWithPrefix(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.New().String())
}

// ValidateWithPrefix checks if a token string has a valid UUID format after the prefix
func ValidateWithPrefix(token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("token is empty")
	}

	parts := strings.Split(token, "_")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid token format: expected prefix_uuid")
	}

	uuidStr := parts[1]
	_, err := uuid.Parse(uuidStr)
	if err != nil {
		return false, fmt.Errorf("invalid UUID format: %w", err)
	}

	return true, nil
}
