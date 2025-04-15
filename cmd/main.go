package main

import (
	app "MydroX/anicetus/internal"
	"MydroX/anicetus/internal/common/jwt"
	cfg "MydroX/anicetus/internal/config"
	"MydroX/anicetus/pkg/cache"
	"MydroX/anicetus/pkg/config"
	"MydroX/anicetus/pkg/db"
	"MydroX/anicetus/pkg/logger"
	"log"

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

	l, err := logger.New(appConfig.Env)
	if err != nil {
		log.Fatalf("error creating logger: %v", err)
	}

	l.Info("connecting to database...")
	connDB, err := db.Connect(appConfig.DB.Host, appConfig.DB.Username, appConfig.DB.Password, appConfig.DB.Name, appConfig.DB.Port)
	if err != nil {
		l.Fatal("error conecting to database", zap.Error(err))
	}
	defer connDB.Close()

	l.Info("creating in memory cache...")
	c, err := cache.New()
	if err != nil {
		l.Fatal("error creating cache", zap.Error(err))
	}

	l.Info("loading allowed audiences in cache...")
	err = jwt.AllowedAudiencesInCache(connDB, c, l)
	if err != nil {
		l.Fatal("error loading allowed audiences in cache", zap.Error(err))
	}

	l.Info("starting server...")
	app.NewServer(&appConfig, l, connDB)
}
