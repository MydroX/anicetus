package logger

import "go.uber.org/zap"

type Logger struct {
	Zap   *zap.Logger
	Debug bool
}

func New(env string) *Logger {
	var zapLogger *zap.Logger
	switch env {
	case "DEV":
		zapLogger, _ = zap.NewDevelopment()
		return &Logger{
			Zap:   zapLogger,
			Debug: true,
		}
	case "PROD":
		zapLogger, _ = zap.NewProduction()
		return &Logger{
			Zap:   zapLogger,
			Debug: false,
		}
	case "TEST":
		zapLogger, _ = zap.NewDevelopment()
		return &Logger{
			Zap:   zapLogger,
			Debug: true,
		}
	default:
		zapLogger = zap.NewNop()
		zapLogger.Warn("logger is set unknown environment")
	}
	return &Logger{
		Zap:   zapLogger,
		Debug: false,
	}
}
