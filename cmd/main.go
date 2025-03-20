package main

import (
	app "MydroX/anicetus/internal"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/pkg/db"
	loggerpkg "MydroX/anicetus/pkg/logger"
	"log"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	logger := loggerpkg.New(cfg.Env)

	connDB, err := db.Connect(cfg.DB.Host, cfg.DB.Username, cfg.DB.Password, cfg.DB.Name, cfg.DB.Port)
	if err != nil {
		logger.Zap.Fatal("error conecting to database", zap.Error(err))
	}
	defer connDB.Close()

	logger.Zap.Info("starting server...")
	app.NewServer(cfg, logger, connDB)
}
