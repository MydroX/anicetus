package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"MydroX/project-v/internal/config"
	iamcontroller "MydroX/project-v/internal/iam/controller"
	iamusecases "MydroX/project-v/internal/iam/usecases"
	userscontroller "MydroX/project-v/internal/users/controller"
	usersrepository "MydroX/project-v/internal/users/repository"
	usersusecases "MydroX/project-v/internal/users/usecases"
	loggerpkg "MydroX/project-v/pkg/logger"
)

type service struct {
	usersController userscontroller.ControllerInterface
	iamController   iamcontroller.ControllerInterface
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

	// - Middleware SECRET KEY API for every endpoint in headers

	api := router.Group("api")
	v1 := api.Group("/v1")

	iamcontroller.Router(v1, service.iamController)
	userscontroller.Router(v1, service.usersController)

	return router
}

// NewServer is a function to start the server for the service.
func NewServer(c *config.Config, logger *loggerpkg.Logger, db *pgxpool.Pool) {
	usersRepository := usersrepository.NewRepository(logger, db)
	// iamRepository := iamrepository.NewRepository(logger, db)

	usersUsecase := usersusecases.NewUsecases(logger, usersRepository, &c.JWT)
	iamUsecase := iamusecases.NewUsecases(logger, usersRepository, &c.JWT)

	usersController := userscontroller.New(logger, usersUsecase, c)
	iamController := iamcontroller.New(logger, iamUsecase, c)

	service := service{
		usersController: usersController,
		iamController:   iamController,
	}

	router := Router(logger, service)

	err := router.Run(fmt.Sprintf(":%s", c.Port))
	if err != nil {
		logger.Zap.Fatal("error starting server", zap.Error(err))
	}
}
