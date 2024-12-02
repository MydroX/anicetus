// Package users is the outside layer of the service, make the link between the API and the internal logic
package users

import (
	"MydroX/project-v/internal/gateway/users/config"
	"MydroX/project-v/internal/gateway/users/dto"
	"MydroX/project-v/internal/gateway/users/models"
	"MydroX/project-v/internal/gateway/users/usecases"
	apiError "MydroX/project-v/pkg/errors"
	"MydroX/project-v/pkg/logger"
	"MydroX/project-v/pkg/response"
	"MydroX/project-v/pkg/uuid"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Controller struct {
	logger   *logger.Logger
	validate *validator.Validate
	usecases usecases.UsersUsecases
	config   *config.Config
}

// NewController is the interface for the controller.
func NewController(l *logger.Logger, u usecases.UsersUsecases, c *config.Config) *Controller {
	validator := validator.New()

	return &Controller{
		validate: validator,
		logger:   l,
		usecases: u,
		config:   c,
	}
}

func (c *Controller) CreateUser(ctx *gin.Context) {
	var request dto.CreateUserRequest

	err := ctx.BindJSON(&request)
	if err != nil {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	// ^[A-Za-z0-9._-]{4,18}$  // username
	usernameRegex, _ := regexp.Compile("[A-Za-z0-9._-]{4,18}$")
	match := usernameRegex.MatchString(request.Username)
	if !match {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	user := models.User{
		Username: request.Username,
		Email:    request.Email,
		Role:     request.Role,
		Password: request.Password,
	}

	err = c.usecases.Create(&user)
	if err != nil {
		response.InternalError(c.logger, ctx, err)
		return
	}

	ctx.JSON(201, gin.H{"message": "user created"})
}

func (c *Controller) GetUser(ctx *gin.Context) {
	userUUID := ctx.Param("uuid")
	if userUUID == "" {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	err := uuid.ValidateWithPrefix(userUUID)
	if err != nil {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	resp, err := c.usecases.Get(userUUID)
	if err != nil {
		if err == apiError.ErrNotFound {
			response.NotFound(c.logger, ctx)
			return
		}
		response.InternalError(c.logger, ctx, err)
		return
	}

	ctx.JSON(200, resp)
}

func (c *Controller) UpdateUser(ctx *gin.Context) {
	var request dto.UpdateUserRequest

	err := ctx.BindJSON(&request)
	if err != nil {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	user := models.User{
		UUID:     request.UUID,
		Username: request.Username,
		Password: request.Password,
		Email:    request.Email,
		Role:     request.Role,
	}

	err = c.usecases.Update(&user)
	if err != nil {
		response.InternalError(c.logger, ctx, err)
		return
	}

	ctx.JSON(200, gin.H{"message": "user updated"})
}

func (c *Controller) UpdateEmail(ctx *gin.Context) {
	var request dto.UpdateEmailRequest

	userUUID := ctx.Param("uuid")
	if userUUID == "" {
		response.InvalidRequest(c.logger, ctx)
		return
	}
	err := uuid.ValidateWithPrefix(userUUID)
	if err != nil {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	err = ctx.BindJSON(&request)
	if err != nil {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	err = c.usecases.UpdateEmail(userUUID, request.Email)
	if err != nil {
		response.InternalError(c.logger, ctx, err)
		return
	}

	ctx.JSON(200, gin.H{"message": "email updated"})
}

func (c *Controller) UpdatePassword(ctx *gin.Context) {
	var request dto.UpdatePasswordRequest

	userUUID := ctx.Param("uuid")
	if userUUID == "" {
		response.InvalidRequest(c.logger, ctx)
		return
	}
	err := uuid.ValidateWithPrefix(userUUID)
	if err != nil {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	err = ctx.BindJSON(&request)
	if err != nil {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	err = c.usecases.UpdatePassword(userUUID, request.Password)

	if err != nil {
		response.InternalError(c.logger, ctx, err)
		return
	}

	ctx.JSON(200, gin.H{"message": "password updated"})
}

func (c *Controller) DeleteUser(ctx *gin.Context) {
	userUUID := ctx.Param("uuid")
	if userUUID == "" {
		response.InvalidRequest(c.logger, ctx)
		return
	}
	err := uuid.ValidateWithPrefix(userUUID)
	if err != nil {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	err = c.usecases.Delete(userUUID)
	if err != nil {
		response.InternalError(c.logger, ctx, err)
		return
	}

	ctx.JSON(200, gin.H{"message": "user deleted"})
}

func (c *Controller) Login(ctx *gin.Context) {
	var request dto.LoginRequest
	if err := ctx.BindJSON(&request); err != nil {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	if request.Username == "" && request.Email == "" {
		response.InvalidRequest(c.logger, ctx)
		return
	}

	token, err := c.usecases.Login(request.Username, request.Email, request.Password)
	if err != nil {
		response.InternalError(c.logger, ctx, err)
		return
	}

	ctx.SetCookie("auth_token", token, 3600, "/", c.config.App.Domain, true, true)

	resp := dto.LoginResponse{
		Token: token,
	}
	ctx.JSON(200, resp)
}
