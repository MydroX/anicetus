package api

import (
	"context"
	"fmt"
	"net/http"

	"MydroX/anicetus/internal/common/jwt"
	"MydroX/anicetus/internal/config"
	iamcontroller "MydroX/anicetus/internal/iam/controller"
	iamrepository "MydroX/anicetus/internal/iam/repository"
	iamusecases "MydroX/anicetus/internal/iam/usecases"
	"MydroX/anicetus/internal/middlewares"
	userscontroller "MydroX/anicetus/internal/users/controller"
	usersrepository "MydroX/anicetus/internal/users/repository"
	usersusecases "MydroX/anicetus/internal/users/usecases"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valkey-io/valkey-go"
	"go.uber.org/zap"
)

type APIServices struct {
	Config *config.Config
	Logger *zap.SugaredLogger
	DB     *pgxpool.Pool
	Valkey valkey.Client
}

type service struct {
	usersController userscontroller.ControllerInterface
	iamController   iamcontroller.ControllerInterface
}

// Router is a function to define the routes for the service.
func Router(logger *zap.SugaredLogger, service service) *gin.Engine {
	router := gin.Default()

	err := router.SetTrustedProxies(nil)
	if err != nil {
		logger.Fatal("error setting trusted proxies", zap.Error(err))
	}

	router.Use(middlewares.TraceID())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
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
func NewServer(s *APIServices) {
	tokenConfig, err := jwt.NewTokenConfigFromEnv(s.Config)
	if err != nil {
		s.Logger.Fatal("error creating token config", zap.Error(err))
	}

	jwtService := jwt.NewJWTService(tokenConfig)

	usersRepository := usersrepository.New(s.Logger, s.DB)
	iamRepository := iamrepository.NewIAMStore(s.Logger, s.DB)
	audienceStore := iamrepository.NewAudienceStore(s.Logger, s.DB)
	audienceManager := iamusecases.NewAudienceManager(s.Logger, audienceStore, s.Valkey)

	if cacheErr := audienceManager.CacheAllowedAudiences(context.Background()); cacheErr != nil {
		s.Logger.Warnw("Failed to prime audience cache", "error", cacheErr)
	}

	usersUsecase := usersusecases.New(s.Logger, usersRepository, s.Config)
	iamUsecase := iamusecases.New(s.Logger, usersRepository, iamRepository, s.Config, jwtService, audienceStore, audienceManager)

	usersController := userscontroller.New(s.Logger, usersUsecase, s.Config)
	iamController := iamcontroller.New(s.Logger, iamUsecase, s.Config)

	service := service{
		usersController: usersController,
		iamController:   iamController,
	}

	router := Router(s.Logger, service)

	err = router.Run(fmt.Sprintf(":%s", s.Config.Port))
	if err != nil {
		s.Logger.Fatal("error starting server", zap.Error(err))
	}
}
