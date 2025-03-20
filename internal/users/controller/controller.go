package controller

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errors"
	"MydroX/anicetus/internal/common/response"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/users/dto"
	"MydroX/anicetus/internal/users/usecases"
	"MydroX/anicetus/pkg/logger"
	"MydroX/anicetus/pkg/password"
	"MydroX/anicetus/pkg/uuid"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	UUID              = "uuid"
	UsernameError     = "username needs to be between 4 and 18 characters long and can only contain letters, numbers, and the following characters: . _ -"
	PasswordError     = "password needs to be between 8 and 32 characters long with at least one uppercase letter, one lowercase letter, one number, and one special character"
	passwordMinLength = 8
	passwordMaxLength = 32
)

type controller struct {
	logger            *logger.Logger
	validate          *validator.Validate
	passwordValidator *password.Validator
	usecases          usecases.UsersUsecases
	config            *config.Config
}

func New(l *logger.Logger, u usecases.UsersUsecases, c *config.Config) ControllerInterface {
	validator := validator.New()

	passwordValidator := password.NewValidator(
		password.WithMinLength(passwordMinLength),
		password.WithMaxLength(passwordMaxLength),
	)

	return &controller{
		validate:          validator,
		logger:            l,
		usecases:          u,
		passwordValidator: passwordValidator,
		config:            c,
	}
}

func (c *controller) CreateUser(ginCtx *gin.Context) {
	var request dto.CreateUserRequest
	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	err := ginCtx.BindJSON(&request)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_FAIL_TO_BIND, Err: err})
		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_INVALID_INPUT, Err: err})
		return
	}

	// ^[A-Za-z0-9._-]{4,18}$  // username
	usernameRegex, _ := regexp.Compile("[A-Za-z0-9._-]{4,18}$")
	match := usernameRegex.MatchString(request.Username)
	if !match {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_INVALID_USERNAME}, response.WithClientMessage(UsernameError))
		return
	}

	err = c.passwordValidator.Validate(request.Password)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_INVALID_PASSWORD}, response.WithLogMessage("failed to validate password: %v", err))
		return
	}

	apiErr := c.usecases.Create(ctx, &request)
	if apiErr != nil {
		if apiErr.Code == errors.ERROR_DUPLICATE_ENTITY {
			response.Conflict(
				c.logger,
				ctx,
				apiErr,
				response.WithLogMessage("failed to create user: %s", err),
			)
			return
		}
		response.InternalError(c.logger, ctx, apiErr)
		return
	}

	ginCtx.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

func (c *controller) GetUser(ginCtx *gin.Context) {
	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	userUUID := ginCtx.Param(UUID)

	err := uuid.ValidateWithPrefix(userUUID)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_INVALID_UUID})
		return
	}

	resp, apiErr := c.usecases.Get(ctx, userUUID)
	if apiErr != nil {
		if apiErr.Code == errors.ERROR_NOT_FOUND {
			response.NotFound(c.logger, ctx, apiErr)
			return
		}
		response.InternalError(c.logger, ctx, apiErr)
		return
	}

	ginCtx.JSON(http.StatusOK, resp)
}

func (c *controller) UpdateUser(ginCtx *gin.Context) {
	var request dto.UpdateUserRequest
	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	err := ginCtx.BindJSON(&request)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_FAIL_TO_BIND, Err: err})
		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_INVALID_INPUT, Err: err})
		return
	}

	apiErr := c.usecases.Update(ctx, &request)
	if apiErr != nil {
		response.InternalError(c.logger, ctx, apiErr)
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "user updated"})
}

func (c *controller) UpdateEmail(ginCtx *gin.Context) {
	var request dto.UpdateEmailRequest
	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	userUUID := ginCtx.Param(UUID)

	err := uuid.ValidateWithPrefix(userUUID)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_INVALID_UUID})
		return
	}

	err = ginCtx.BindJSON(&request)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_FAIL_TO_BIND, Err: err})
		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_INVALID_INPUT, Err: err})
		return
	}

	apiErr := c.usecases.UpdateEmail(ctx, userUUID, request.Email)
	if apiErr != nil {
		response.InternalError(c.logger, ctx, apiErr)
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "email updated"})
}

func (c *controller) UpdatePassword(ginCtx *gin.Context) {
	var request dto.UpdatePasswordRequest
	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	userUUID := ginCtx.Param(UUID)

	err := uuid.ValidateWithPrefix(userUUID)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_INVALID_UUID})
		return
	}

	err = ginCtx.BindJSON(&request)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_FAIL_TO_BIND, Err: err})
		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_INVALID_INPUT, Err: err})
		return
	}

	err = c.passwordValidator.Validate(request.Password)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_INVALID_PASSWORD}, response.WithLogMessage("failed to validate password: %v", err))
		return
	}

	apiErr := c.usecases.UpdatePassword(ctx, userUUID, request.Password)
	if apiErr != nil {
		response.InternalError(c.logger, ctx, apiErr)
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

func (c *controller) DeleteUser(ginCtx *gin.Context) {
	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	userUUID := ginCtx.Param(UUID)

	err := uuid.ValidateWithPrefix(userUUID)
	if err != nil {
		response.BadRequest(c.logger, ctx, &errors.Err{Code: errors.ERROR_INVALID_UUID})
		return
	}

	apiErr := c.usecases.Delete(ctx, userUUID)
	if apiErr != nil {
		response.InternalError(c.logger, ctx, apiErr)
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func (c *controller) GetAllUsers(ginCtx *gin.Context) {
	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	resp, err := c.usecases.GetAllUsers(ctx)
	if err != nil {
		response.InternalError(c.logger, ctx, err)
		return
	}

	ginCtx.JSON(http.StatusOK, resp)
}
