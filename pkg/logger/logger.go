package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func New(env string) (*zap.SugaredLogger, error) {
	var zapLogger *zap.Logger
	var err error

	switch env {
	case "DEV":
		zapLogger, err = zap.NewDevelopment()
		if err != nil {
			return nil, fmt.Errorf("failed to create development logger: %w", err)
		}
	case "PROD":
		zapLogger, err = zap.NewProduction()
		if err != nil {
			return nil, fmt.Errorf("failed to create production logger: %w", err)
		}
	case "TEST":
		zapLogger, err = zap.NewDevelopment()
		if err != nil {
			return nil, fmt.Errorf("failed to create test logger: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid environment for logger: %s", env)
	}

	return zapLogger.Sugar(), nil
}
