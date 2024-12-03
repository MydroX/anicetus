package uuid

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func NewWithPrefix(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, uuid.New().String())
}

func ValidateWithPrefix(uuidWithPrefix string) error {
	str := strings.Split(uuidWithPrefix, "-")
	if len(str) < 1 {
		return fmt.Errorf("invalid uuid")
	}

	uuidSliced := str[1:]
	uuidStr := strings.Join(uuidSliced, "-")

	err := uuid.Validate(uuidStr)
	if err != nil {
		return err
	}
	return nil
}
