package main

import (
	"MydroX/project-v/internal/gateway"
	"MydroX/project-v/internal/gateway/users/config"
	"MydroX/project-v/pkg/db"
	loggerpkg "MydroX/project-v/pkg/logger"
	"log"

	"go.uber.org/zap"
)

const serviceName = "gateway"

func main() {
	cfg, err := config.LoadConfig(serviceName)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	logger := loggerpkg.New(cfg.Env)

	database, err := db.Connect(cfg.DB.Host, cfg.DB.Username, cfg.DB.Password, cfg.DB.Name, cfg.DB.Port)
	if err != nil {
		logger.Zap.Fatal("error conecting to database", zap.Error(err))
	}

	logger.Zap.Info("starting server...")
	gateway.NewServer(cfg, logger, database)
}
