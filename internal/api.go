package api

import (
	"MydroX/project-v/internal/config"
	usersservice "MydroX/project-v/internal/users"
	"MydroX/project-v/internal/users/repository"
	"MydroX/project-v/internal/users/usecases"
	loggerpkg "MydroX/project-v/pkg/logger"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type service struct {
	usersController *usersservice.Controller
}

// Router is a function to define the routes for the service.
func Router(logger *loggerpkg.Logger, service service) *gin.Engine {
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

	// Public routes
	users.POST("/register", service.usersController.CreateUser)
	users.POST("/login", service.usersController.Login)

	// Logged in routes
	users.PUT("/:uuid", service.usersController.UpdateUser)
	users.PATCH("/:uuid/email", service.usersController.UpdateEmail)
	users.PATCH("/:uuid/password", service.usersController.UpdatePassword)

	// Admin routes
	users.GET("/:uuid", service.usersController.GetUser)
	users.DELETE("/:uuid", service.usersController.DeleteUser)

	return router
}

// NewServer is a function to start the server for the service.
func NewServer(c *config.Config, logger *loggerpkg.Logger, db *pgxpool.Pool) {
	usersRepository := repository.NewRepository(logger, db)

	usersUsecase := usecases.NewUsecases(logger, usersRepository, &c.JWT)

	usersController := usersservice.NewController(logger, usersUsecase, c)

	service := service{
		usersController: usersController,
	}

	router := Router(logger, service)

	err := router.Run(fmt.Sprintf(":%s", c.Port))
	if err != nil {
		logger.Zap.Fatal("error starting server", zap.Error(err))
	}
}
