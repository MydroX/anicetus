package iam

import (
	"MydroX/project-v/internal/common/errorscode"
	"MydroX/project-v/internal/common/response"
	"MydroX/project-v/internal/config"
	"MydroX/project-v/internal/iam/dto"
	"MydroX/project-v/internal/iam/usecases"
	"MydroX/project-v/pkg/logger"
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
	ctx := ginCtx.Request.Context()

	if err := ginCtx.BindJSON(&request); err != nil {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_REQUEST)
		return
	}

	err := c.validate.Struct(request)
	if err != nil {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_REQUEST)
		return
	}

	if request.Username == "" && request.Email == "" {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_REQUEST)
		return
	}

	accessToken, refreshToken, err := c.usecases.Login(&ctx, request.Username, request.Email, request.Password)
	if err != nil {
		if ctx.Value(errorscode.CtxErrorCodeKey) == errorscode.CODE_ENTITY_NOT_FOUND {
			response.NotFound(c.logger, ginCtx, errorscode.CODE_ENTITY_NOT_FOUND)
			return
		}
		response.InternalError(c.logger, ginCtx, err, ginCtx.GetString(string(errorscode.CtxErrorCodeKey)))
		return
	}

	ginCtx.SetCookie("access_token", accessToken, c.config.JWT.AccessToken.Expiration, "/", c.config.App.Domain, true, true)
	ginCtx.SetCookie("refresh_token", refreshToken, c.config.JWT.RefreshToken.Expiration, "/", c.config.App.Domain, true, true)

	resp := dto.LoginResponse{
		Token: accessToken,
	}
	ginCtx.JSON(http.StatusOK, resp)
}

// nolint
func (c *controller) RefreshToken(_ *gin.Context) {
	return
}
