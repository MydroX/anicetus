package api

import (
	"fmt"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"MydroX/anicetus/internal/common/jwt"
	"MydroX/anicetus/internal/config"
	iamcontroller "MydroX/anicetus/internal/iam/controller"
	iamrepository "MydroX/anicetus/internal/iam/repository"
	iamusecases "MydroX/anicetus/internal/iam/usecases"
	userscontroller "MydroX/anicetus/internal/users/controller"
	usersrepository "MydroX/anicetus/internal/users/repository"
	usersusecases "MydroX/anicetus/internal/users/usecases"
)

type APIServices struct {
	Config        *config.Config
	Logger        *zap.SugaredLogger
	DB            *pgxpool.Pool
	CacheInMemory *ristretto.Cache[string, string]
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
func NewServer(s *APIServices) {
	// audienceManager := jwt.NewAudienceManager(s.Logger, s.DB, s.CacheInMemory)

	jwtService := jwt.NewJWTService(jwt.TokenConfig{
		SecretKey:        s.Config.JWT.Secret,
		ClockSkewSeconds: s.Config.JWT.SkewSeconds,
		ExpectedIssuer:   s.Config.JWT.Issuer,
	})

	usersRepository := usersrepository.New(s.Logger, s.DB)
	iamRepository := iamrepository.NewIAMStore(s.Logger, s.DB)

	usersUsecase := usersusecases.New(s.Logger, usersRepository, s.Config)
	iamUsecase := iamusecases.New(s.Logger, usersRepository, iamRepository, s.Config, jwtService)

	usersController := userscontroller.New(s.Logger, usersUsecase, s.Config)
	iamController := iamcontroller.New(s.Logger, iamUsecase, s.Config)

	service := service{
		usersController: usersController,
		iamController:   iamController,
	}

	router := Router(s.Logger, service)

	err := router.Run(fmt.Sprintf(":%s", s.Config.Port))
	if err != nil {
		s.Logger.Fatal("error starting server", zap.Error(err))
	}
}
