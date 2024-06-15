package uuid

import (
	"github.com/google/uuid"
)

func ValidateAndParse(uuidStr string) (uuid.UUID, error) {
	err := uuid.Validate(uuidStr)
	if err != nil {
		return uuid.UUID{}, err
	}

	userUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.UUID{}, err
	}

	return userUUID, nil
}
