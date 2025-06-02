package iam

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/common/response"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/iam/dto"
	"MydroX/anicetus/internal/iam/usecases"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type controller struct {
	logger   *zap.SugaredLogger
	validate *validator.Validate
	usecases usecases.IamUsecasesService
	config   *config.Config
}

func New(l *zap.SugaredLogger, u usecases.IamUsecasesService, c *config.Config) ControllerInterface {
	v := validator.New()

	return &controller{
		validate: v,
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
		response.BadRequest(c.logger, ctx, errorsutil.ERROR_FAIL_TO_BIND, errorsutil.MessageFailToBind)
		return
	}

	jwtErr := c.validate.Struct(request)
	if jwtErr != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ERROR_INVALID_INPUT, errorsutil.MessageInvalidInput)
		return
	}

	if request.Username == "" && request.Email == "" {
		response.BadRequest(c.logger, ctx, errorsutil.ERROR_INVALID_INPUT, "username or email is required")
		return
	}

	accessToken, refreshToken, jwtErr := c.usecases.Login(ctx, &request)
	if jwtErr != nil {
		response.Error(c.logger, ctx, jwtErr)
		return
	}

	ginCtx.SetCookie("access_token", accessToken, c.config.JWT.AccessToken.Expiration, "/", c.config.App.Domain, true, true)
	ginCtx.SetCookie("refresh_token", refreshToken, c.config.JWT.RefreshToken.Expiration, "/", c.config.App.Domain, true, true)

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
