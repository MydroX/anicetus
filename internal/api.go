package api

import (
	"context"
	"fmt"
	"net/http"

	authcontroller "MydroX/anicetus/internal/authentication/controller"
	authmiddleware "MydroX/anicetus/internal/authentication/middleware"
	authrepository "MydroX/anicetus/internal/authentication/repository"
	authusecases "MydroX/anicetus/internal/authentication/usecases"
	authzcontroller "MydroX/anicetus/internal/authorization/controller"
	authzrepository "MydroX/anicetus/internal/authorization/repository"
	authzusecases "MydroX/anicetus/internal/authorization/usecases"
	"MydroX/anicetus/internal/config"
	identitycontroller "MydroX/anicetus/internal/identity/controller"
	identityrepository "MydroX/anicetus/internal/identity/repository"
	identityusecases "MydroX/anicetus/internal/identity/usecases"
	"MydroX/anicetus/internal/middlewares"
	servicescontroller "MydroX/anicetus/internal/services/controller"
	servicesrepository "MydroX/anicetus/internal/services/repository"
	servicesusecases "MydroX/anicetus/internal/services/usecases"
	"MydroX/anicetus/pkg/jwt"
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
	identityController       identitycontroller.ControllerInterface
	authenticationController authcontroller.ControllerInterface
	authorizationController  authzcontroller.ControllerInterface
	servicesController       servicescontroller.ControllerInterface
}

// Router is a function to define the routes for the service.
func Router(logger *zap.SugaredLogger, svc service, jwtService *jwt.Service) *gin.Engine {
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

	api := router.Group("api")
	v1 := api.Group("/v1")

	// Public routes
	authcontroller.PublicRouter(v1, svc.authenticationController)
	identitycontroller.Router(v1, svc.identityController)

	// Authenticated routes
	authenticated := v1.Group("")
	authenticated.Use(authmiddleware.AuthMiddleware(jwtService))

	authcontroller.AuthenticatedRouter(authenticated, svc.authenticationController)
	authzcontroller.Router(authenticated, svc.authorizationController)
	servicescontroller.Router(authenticated, svc.servicesController)

	return router
}

// NewServer is a function to start the server for the service.
func NewServer(s *APIServices) {
	tokenConfig, err := jwt.NewTokenConfigFromEnv(s.Config)
	if err != nil {
		s.Logger.Fatal("error creating token config", zap.Error(err))
	}

	jwtService := jwt.NewJWTService(tokenConfig)

	// Repositories
	identityRepository := identityrepository.New(s.Logger, s.DB)
	sessionStore := authrepository.New(s.Logger, s.DB)
	serviceStore := servicesrepository.New(s.Logger, s.DB)

	// Service Manager (Valkey caching)
	serviceManager := servicesusecases.NewServiceManager(s.Logger, serviceStore, s.Valkey)

	if cacheErr := serviceManager.CacheAllowedServices(context.Background()); cacheErr != nil {
		s.Logger.Warnw("Failed to prime services cache", "error", cacheErr)
	}

	// Repositories
	authzRepository := authzrepository.New(s.Logger, s.DB)

	// Usecases
	identityUsecase := identityusecases.New(s.Logger, identityRepository, s.Config)
	authUsecase := authusecases.New(s.Logger, identityRepository, sessionStore, s.Config, jwtService, serviceManager)
	authzUsecase := authzusecases.New(s.Logger, authzRepository)
	servicesUsecase := servicesusecases.New(s.Logger, serviceStore, serviceManager)

	// Controllers
	identityController := identitycontroller.New(s.Logger, identityUsecase, s.Config)
	authController := authcontroller.New(s.Logger, authUsecase, s.Config)
	authzController := authzcontroller.New(s.Logger, authzUsecase)
	servicesController := servicescontroller.New(s.Logger, servicesUsecase, s.Config)

	svc := service{
		identityController:       identityController,
		authenticationController: authController,
		authorizationController:  authzController,
		servicesController:       servicesController,
	}

	router := Router(s.Logger, svc, jwtService)

	err = router.Run(fmt.Sprintf(":%s", s.Config.Port))
	if err != nil {
		s.Logger.Fatal("error starting server", zap.Error(err))
	}
}
