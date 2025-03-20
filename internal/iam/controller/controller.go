package iam

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errors"
	"MydroX/anicetus/internal/common/response"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/iam/dto"
	"MydroX/anicetus/internal/iam/usecases"
	"MydroX/anicetus/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type controller struct {
	logger   *logger.Logger
	validate *validator.Validate
	usecases usecases.IamUsecasesInterface
	config   *config.Config
}

func New(l *logger.Logger, u usecases.IamUsecasesInterface, c *config.Config) ControllerInterface {
	validator := validator.New()

	return &controller{
		validate: validator,
		logger:   l,
		usecases: u,
		config:   c,
	}
}

func (c *controller) Login(ginCtx *gin.Context) {
	var request dto.LoginRequest
	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	if err := ginCtx.BindJSON(&request); err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_FAIL_TO_BIND, Err: err})
		return
	}

	err := c.validate.Struct(request)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_INVALID_INPUT, Err: err})
		return
	}

	if request.Username == "" && request.Email == "" {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_INVALID_INPUT, Message: "username or email is required"})
		return
	}

	accessToken, refreshToken, apiErr := c.usecases.Login(ctx, &request)
	if apiErr != nil {
		if apiErr.Code == errors.ERROR_NOT_FOUND {
			response.NotFound(c.logger, ctx, &errors.Err{Code: errors.ERROR_NOT_FOUND, Message: "user not found"})
			return
		}
		response.InternalError(c.logger, ctx, apiErr)
		return
	}

	ginCtx.SetCookie("access_token", accessToken, c.config.Session.AccessToken.Expiration, "/", c.config.App.Domain, true, true)
	ginCtx.SetCookie("refresh_token", refreshToken, c.config.Session.RefreshToken.Expiration, "/", c.config.App.Domain, true, true)

	resp := dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	ginCtx.JSON(http.StatusOK, resp)
}

// nolint
func (c *controller) RefreshToken(_ *gin.Context) {
	return
}
