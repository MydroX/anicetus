// Package iam is the entry point for the IAM service. It starts the server and defines the routes for the service.
package iam

import (
	"MydroX/project-v/internal/iam/repository"
	"MydroX/project-v/internal/iam/usecases"
	"MydroX/project-v/pkg/logger"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Router(logger *logger.Logger, validate *validator.Validate, db *gorm.DB, controller ControllerInterface) *gin.Engine {
	router := gin.Default()

	err := router.SetTrustedProxies(nil)
	if err != nil {
		logger.Zap.Fatal("error setting trusted proxies", zap.Error(err))
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	api := router.Group("api")

	// - Middleware SECRET KEY API for every endpoint in headers

	v1 := api.Group("/v1")
	v1.POST("/register", controller.CreateUser)
	v1.POST("/auth", controller.AuthenticateUser)
	v1.POST("/:uuid", controller.GetUser)

	// TODO
	// - Middleware authentification
	// - UpdateUser
	// - DeleteUser

	return router
}

// NewServer is a function to start the server for the IAM service.
func NewServer(config *Config, logger *logger.Logger, validate *validator.Validate, db *gorm.DB) {
	repository := repository.NewRepository(logger, db)
	usecases := usecases.NewUsecases(logger, repository)
	controller := NewController(logger, validate, usecases)

	router := Router(logger, validate, db, controller)

	err := router.Run(fmt.Sprintf(":%s", config.Port))
	if err != nil {
		logger.Zap.Fatal("error starting server", zap.Error(err))
	}
}
