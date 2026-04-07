package controller

import (
	"net/http"

	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/services/dto"
	"MydroX/anicetus/internal/services/usecases"
	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/httpresponse"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type controller struct {
	logger   *zap.SugaredLogger
	validate *validator.Validate
	usecases usecases.ServicesUsecases
	config   *config.Config
}

func New(l *zap.SugaredLogger, u usecases.ServicesUsecases, c *config.Config) ControllerInterface {
	v := validator.New()

	return &controller{
		validate: v,
		logger:   l,
		usecases: u,
		config:   c,
	}
}

func (c *controller) RegisterService(ginCtx *gin.Context) {
	var request dto.RegisterServiceRequest

	if err := ginCtx.BindJSON(&request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorFailToBind, errs.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, errs.MessageInvalidInput)

		return
	}

	if err := c.usecases.RegisterService(ginCtx.Request.Context(), &request); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusCreated, gin.H{"message": "service registered"})
}

func (c *controller) RevokeService(ginCtx *gin.Context) {
	service := ginCtx.Param("service")
	if service == "" {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, "service is required")

		return
	}

	if err := c.usecases.RevokeService(ginCtx.Request.Context(), service); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "service revoked"})
}

func (c *controller) GetAllServices(ginCtx *gin.Context) {
	services, err := c.usecases.GetAllServices(ginCtx.Request.Context())
	if err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, dto.ServiceListResponse{Services: services})
}

func (c *controller) GetUserServices(ginCtx *gin.Context) {
	userUUID := ginCtx.Param("uuid")

	if _, err := uuid.Parse(userUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	services, err := c.usecases.GetUserServices(ginCtx.Request.Context(), userUUID)
	if err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, dto.ServiceListResponse{Services: services})
}

func (c *controller) AssignServiceToUser(ginCtx *gin.Context) {
	var request dto.AssignServiceRequest

	userUUID := ginCtx.Param("uuid")

	if _, err := uuid.Parse(userUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	if err := ginCtx.BindJSON(&request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorFailToBind, errs.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, errs.MessageInvalidInput)

		return
	}

	if err := c.usecases.AssignServiceToUser(ginCtx.Request.Context(), userUUID, &request); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "service assigned to user"})
}
