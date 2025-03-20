package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"MydroX/anicetus/internal/config"
	iamcontroller "MydroX/anicetus/internal/iam/controller"
	iamrepository "MydroX/anicetus/internal/iam/repository"
	iamusecases "MydroX/anicetus/internal/iam/usecases"
	userscontroller "MydroX/anicetus/internal/users/controller"
	usersrepository "MydroX/anicetus/internal/users/repository"
	usersusecases "MydroX/anicetus/internal/users/usecases"
	loggerpkg "MydroX/anicetus/pkg/logger"
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
	usersRepository := usersrepository.New(logger, db)
	iamRepository := iamrepository.New(logger, db)

	usersUsecase := usersusecases.New(logger, usersRepository, &c.Session)
	iamUsecase := iamusecases.New(logger, usersRepository, iamRepository, &c.Session)

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
