package controller

import (
	"net/http"

	"MydroX/anicetus/internal/authentication/dto"
	"MydroX/anicetus/internal/authentication/usecases"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/httpresponse"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type controller struct {
	logger   *zap.SugaredLogger
	validate *validator.Validate
	usecases usecases.AuthenticationUsecases
	config   *config.Config
}

func New(l *zap.SugaredLogger, u usecases.AuthenticationUsecases, c *config.Config) ControllerInterface {
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

	if err := ginCtx.BindJSON(&request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorFailToBind, errs.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, errs.MessageInvalidInput)

		return
	}

	if request.Username == "" && request.Email == "" {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, "username or email is required")

		return
	}

	accessToken, refreshToken, err := c.usecases.Login(ginCtx.Request.Context(), &request)
	if err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

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

func (c *controller) Logout(ginCtx *gin.Context) {
	refreshToken, err := ginCtx.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, "refresh token is required")

		return
	}

	if err := c.usecases.Logout(ginCtx.Request.Context(), refreshToken); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	// Clear cookies
	ginCtx.SetCookie("access_token", "", -1, "/", c.config.App.Domain, true, true)
	ginCtx.SetCookie("refresh_token", "", -1, "/", c.config.App.Domain, true, true)

	ginCtx.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (c *controller) RefreshToken(ginCtx *gin.Context) {
	refreshToken, err := ginCtx.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, "refresh token is required")

		return
	}

	newAccessToken, newRefreshToken, refreshErr := c.usecases.RefreshToken(ginCtx.Request.Context(), refreshToken)
	if refreshErr != nil {
		httpresponse.Error(c.logger, ginCtx, refreshErr)

		return
	}

	ginCtx.SetCookie("access_token", newAccessToken, c.config.JWT.AccessToken.Expiration, "/", c.config.App.Domain, true, true)
	ginCtx.SetCookie("refresh_token", newRefreshToken, c.config.JWT.RefreshToken.Expiration, "/", c.config.App.Domain, true, true)

	resp := dto.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}
	ginCtx.JSON(http.StatusOK, resp)
}
