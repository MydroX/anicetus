// Package iam is the entry point for the IAM service. It starts the server and defines the routes for the service.
package iam

import (
	"MydroX/project-v/internal/iam/users"
	"MydroX/project-v/internal/iam/users/repository"
	"MydroX/project-v/internal/iam/users/usecases"
	"MydroX/project-v/pkg/logger"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type service struct {
	usersController *users.Controller
}

// Router is a function to define the routes for the IAM service.
func Router(logger *logger.Logger, service service) *gin.Engine {
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

	users := v1.Group("/users")
	users.POST("/register", service.usersController.CreateUser)
	users.POST("/auth", service.usersController.AuthenticateUser)
	users.GET("/:uuid", service.usersController.GetUser)

	// TODO: Middleware authentification
	users.PUT("/:uuid", service.usersController.UpdateUser)
	users.PATCH("/:uuid/email", service.usersController.UpdateEmail)
	users.PATCH("/:uuid/password", service.usersController.UpdatePassword)

	users.DELETE("/:uuid", service.usersController.DeleteUser)

	return router
}

// NewServer is a function to start the server for the IAM service.
func NewServer(config *Config, logger *logger.Logger, db *gorm.DB) {
	usersRepository := repository.NewRepository(logger, db)

	usersUsecase := usecases.NewUsecases(logger, usersRepository)

	usersController := users.NewController(logger, usersUsecase)

	service := service{
		usersController: usersController,
	}

	router := Router(logger, service)

	err := router.Run(fmt.Sprintf(":%s", config.Port))
	if err != nil {
		logger.Zap.Fatal("error starting server", zap.Error(err))
	}
}
