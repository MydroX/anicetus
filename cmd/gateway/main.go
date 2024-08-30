package main

import (
	"MydroX/project-v/internal/gateway"
	"MydroX/project-v/pkg/db"
	"MydroX/project-v/pkg/logger"
	"log"
)

const serviceName = "gateway"

func main() {
	cfg, err := gateway.LoadConfig(serviceName)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	logger := logger.New(cfg.Env)

	db := db.Connect(cfg.DB.Host, cfg.DB.Username, cfg.DB.Password, cfg.DB.Name, cfg.DB.Port)

	logger.Zap.Info("starting server...")
	gateway.NewServer(cfg, logger, db)
}
