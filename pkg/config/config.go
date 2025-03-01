package config

import (
	"fmt"
	"os"
)

func Read() ([]byte, error) {
	file, err := os.ReadFile("cmd/config.yml")
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}

	return file, nil
}
