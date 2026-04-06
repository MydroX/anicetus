package controller

import (
	"net/http"
	"regexp"

	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/common/response"
	"MydroX/anicetus/internal/config"
	"MydroX/anicetus/internal/users/dto"
	"MydroX/anicetus/internal/users/usecases"
	"MydroX/anicetus/pkg/password"
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

//revive:disable:line-length-limit
const (
	UUID              = "uuid"
	UsernameError     = "username needs to be between 4 and 18 characters long and can only contain letters, numbers, and the following characters: . _ -"
	PasswordError     = "password needs to be between 8 and 32 characters long with at least one uppercase letter, one lowercase letter, one number, and one special character"
	passwordMinLength = 8
	passwordMaxLength = 32
)

type controller struct {
	logger            *zap.SugaredLogger
	validate          *validator.Validate
	passwordValidator *password.Validator
	usecases          usecases.UsersUsecases
	config            *config.Config
}

func New(l *zap.SugaredLogger, u usecases.UsersUsecases, c *config.Config) ControllerInterface {
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
		response.BadRequest(c.logger, ctx, errorsutil.ErrorFailToBind, errorsutil.MessageFailToBind)

		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidInput, errorsutil.MessageInvalidInput)

		return
	}

	// ^[A-Za-z0-9._-]{4,18}$  // username
	usernameRegex, _ := regexp.Compile("[A-Za-z0-9._-]{4,18}$")

	match := usernameRegex.MatchString(request.Username)
	if !match {
		// TODO: Need to change the message depending on the parameters set in rules
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidUsername, UsernameError)

		return
	}

	err = c.passwordValidator.Validate(request.Password)
	if err != nil {
		// TODO: Need to change the message depending on the parameters set in rules
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidPassword, PasswordError)

		return
	}

	err = c.usecases.Create(ctx, &request)
	if err != nil {
		response.Error(c.logger, ctx, err)

		return
	}

	ginCtx.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

func (c *controller) GetUser(ginCtx *gin.Context) {
	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	userUUID := ginCtx.Param(UUID)

	if _, err := uuid.Parse(userUUID); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidUUID, errorsutil.MessageInvalidUUID)

		return
	}

	resp, err := c.usecases.Get(ctx, userUUID)
	if err != nil {
		response.Error(c.logger, ctx, err)

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
		response.BadRequest(c.logger, ctx, errorsutil.ErrorFailToBind, errorsutil.MessageFailToBind)

		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidInput, errorsutil.MessageInvalidInput)

		return
	}

	apiErr := c.usecases.Update(ctx, &request)
	if apiErr != nil {
		response.Error(c.logger, ctx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "user updated"})
}

func (c *controller) UpdateEmail(ginCtx *gin.Context) {
	var request dto.UpdateEmailRequest

	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	userUUID := ginCtx.Param(UUID)

	if _, err := uuid.Parse(userUUID); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidUUID, errorsutil.MessageInvalidUUID)

		return
	}

	if err := ginCtx.BindJSON(&request); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorFailToBind, errorsutil.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidInput, errorsutil.MessageInvalidInput)

		return
	}

	apiErr := c.usecases.UpdateEmail(ctx, userUUID, request.Email)
	if apiErr != nil {
		response.Error(c.logger, ctx, apiErr)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "email updated"})
}

func (c *controller) UpdatePassword(ginCtx *gin.Context) {
	var request dto.UpdatePasswordRequest

	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	userUUID := ginCtx.Param(UUID)

	if _, err := uuid.Parse(userUUID); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidUUID, errorsutil.MessageInvalidUUID)

		return
	}

	if err := ginCtx.BindJSON(&request); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorFailToBind, errorsutil.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidInput, errorsutil.MessageInvalidInput)

		return
	}

	if err := c.passwordValidator.Validate(request.Password); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidPassword, PasswordError)

		return
	}

	if err := c.usecases.UpdatePassword(ctx, userUUID, request.Password); err != nil {
		response.Error(c.logger, ctx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

func (c *controller) DeleteUser(ginCtx *gin.Context) {
	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	userUUID := ginCtx.Param(UUID)

	if _, err := uuid.Parse(userUUID); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidUUID, errorsutil.MessageInvalidUUID)

		return
	}

	apiErr := c.usecases.Delete(ctx, userUUID)
	if apiErr != nil {
		response.Error(c.logger, ctx, apiErr)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func (c *controller) GetAllUsers(ginCtx *gin.Context) {
	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	resp, err := c.usecases.GetAllUsers(ctx)
	if err != nil {
		response.Error(c.logger, ctx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, resp)
}
