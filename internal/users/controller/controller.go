package controller

import (
	"net/http"
	"regexp"

	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/httpresponse"
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
	v := validator.New()

	passwordValidator := password.NewValidator(
		password.WithMinLength(passwordMinLength),
		password.WithMaxLength(passwordMaxLength),
	)

	return &controller{
		validate:          v,
		logger:            l,
		usecases:          u,
		passwordValidator: passwordValidator,
		config:            c,
	}
}

func (c *controller) CreateUser(ginCtx *gin.Context) {
	var request dto.CreateUserRequest


	err := ginCtx.BindJSON(&request)
	if err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorFailToBind, errs.MessageFailToBind)

		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, errs.MessageInvalidInput)

		return
	}

	// ^[A-Za-z0-9._-]{4,18}$  // username
	usernameRegex, _ := regexp.Compile("[A-Za-z0-9._-]{4,18}$")

	match := usernameRegex.MatchString(request.Username)
	if !match {
		// TODO: Need to change the message depending on the parameters set in rules
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUsername, UsernameError)

		return
	}

	err = c.passwordValidator.Validate(request.Password)
	if err != nil {
		// TODO: Need to change the message depending on the parameters set in rules
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidPassword, PasswordError)

		return
	}

	err = c.usecases.Create(ginCtx.Request.Context(), &request)
	if err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

func (c *controller) GetUser(ginCtx *gin.Context) {

	userUUID := ginCtx.Param(UUID)

	if _, err := uuid.Parse(userUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	resp, err := c.usecases.Get(ginCtx.Request.Context(), userUUID)
	if err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, resp)
}

func (c *controller) UpdateUser(ginCtx *gin.Context) {
	var request dto.UpdateUserRequest


	err := ginCtx.BindJSON(&request)
	if err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorFailToBind, errs.MessageFailToBind)

		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, errs.MessageInvalidInput)

		return
	}

	apiErr := c.usecases.Update(ginCtx.Request.Context(), &request)
	if apiErr != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "user updated"})
}

func (c *controller) UpdateEmail(ginCtx *gin.Context) {
	var request dto.UpdateEmailRequest


	userUUID := ginCtx.Param(UUID)

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

	apiErr := c.usecases.UpdateEmail(ginCtx.Request.Context(), userUUID, request.Email)
	if apiErr != nil {
		httpresponse.Error(c.logger, ginCtx, apiErr)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "email updated"})
}

func (c *controller) UpdatePassword(ginCtx *gin.Context) {
	var request dto.UpdatePasswordRequest


	userUUID := ginCtx.Param(UUID)

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

	if err := c.passwordValidator.Validate(request.Password); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidPassword, PasswordError)

		return
	}

	if err := c.usecases.UpdatePassword(ginCtx.Request.Context(), userUUID, request.Password); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

func (c *controller) DeleteUser(ginCtx *gin.Context) {

	userUUID := ginCtx.Param(UUID)

	if _, err := uuid.Parse(userUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	apiErr := c.usecases.Delete(ginCtx.Request.Context(), userUUID)
	if apiErr != nil {
		httpresponse.Error(c.logger, ginCtx, apiErr)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func (c *controller) GetAllUsers(ginCtx *gin.Context) {

	resp, err := c.usecases.GetAllUsers(ginCtx.Request.Context())
	if err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, resp)
}
