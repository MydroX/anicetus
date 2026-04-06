package main

import (
	"log"

	app "MydroX/anicetus/internal"
	cfg "MydroX/anicetus/internal/config"
	valkeyCache "MydroX/anicetus/pkg/cache/valkey"
	"MydroX/anicetus/pkg/config"
	"MydroX/anicetus/pkg/db"
	"MydroX/anicetus/pkg/logger"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	var appConfig cfg.Config

	err = viper.Unmarshal(&appConfig)
	if err != nil {
		log.Fatalf("error unmarshalling config: %v", err)
	}

	if err := appConfig.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	l, err := logger.New(appConfig.Env)
	if err != nil {
		log.Fatalf("error creating logger: %v", err)
	}

	l.Info("connecting to database...")

	connDB, err := db.Connect(
		appConfig.Database.Host,
		appConfig.Database.Username,
		appConfig.Database.Password,
		appConfig.Database.Name,
		appConfig.Database.Port,
	)
	if err != nil {
		l.Fatal("error connecting to database", zap.Error(err))
	}

	defer connDB.Close()

	l.Info("connecting to valkey...")

	valkeyClient, err := valkeyCache.NewClient(appConfig.Valkey.Address)
	if err != nil {
		l.Fatal("error connecting to valkey", zap.Error(err))
	}

	defer valkeyClient.Close()

	l.Info("starting server...")
	app.NewServer(
		&app.APIServices{
			Config: &appConfig,
			Logger: l,
			DB:     connDB,
			Valkey: valkeyClient,
		},
	)
}
