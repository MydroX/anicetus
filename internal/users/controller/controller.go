package controller

import (
	"MydroX/project-v/internal/common/errorscode"
	"MydroX/project-v/internal/common/response"
	"MydroX/project-v/internal/config"
	"MydroX/project-v/internal/users/dto"
	"MydroX/project-v/internal/users/usecases"
	"MydroX/project-v/pkg/logger"
	"MydroX/project-v/pkg/uuid"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const UUID = "uuid"

type controller struct {
	logger   *logger.Logger
	validate *validator.Validate
	usecases usecases.UsersUsecases
	config   *config.Config
}

func New(l *logger.Logger, u usecases.UsersUsecases, c *config.Config) ControllerInterface {
	validator := validator.New()

	return &controller{
		validate: validator,
		logger:   l,
		usecases: u,
		config:   c,
	}
}

func (c *controller) CreateUser(ginCtx *gin.Context) {
	var request dto.CreateUserRequest
	ctx := ginCtx.Request.Context()

	err := ginCtx.BindJSON(&request)
	if err != nil {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_REQUEST)
		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_REQUEST)
		return
	}

	// ^[A-Za-z0-9._-]{4,18}$  // username
	usernameRegex, _ := regexp.Compile("[A-Za-z0-9._-]{4,18}$")
	match := usernameRegex.MatchString(request.Username)
	if !match {
		response.BadRequestWithMessage(c.logger, ginCtx, errorscode.CODE_INVALID_USERNAME, "invalid username")
		return
	}

	err = c.usecases.Create(&ctx, &request)
	if err != nil {
		if ctx.Value(errorscode.CtxErrorCodeKey) == errorscode.CODE_DUPLICATE_ENTITY {
			response.Conflict(c.logger, ginCtx, errorscode.CODE_DUPLICATE_ENTITY)
			return
		}
		response.InternalError(c.logger, ginCtx, err, ginCtx.GetString(string(errorscode.CtxErrorCodeKey)))
		return
	}

	ginCtx.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

func (c *controller) GetUser(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()

	userUUID := ginCtx.Param(UUID)

	err := uuid.ValidateWithPrefix(userUUID)
	if err != nil {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_UUID)
		return
	}

	resp, err := c.usecases.Get(&ctx, userUUID)
	if err != nil {
		if ctx.Value(errorscode.CtxErrorCodeKey) == errorscode.CODE_ENTITY_NOT_FOUND {
			response.NotFound(c.logger, ginCtx, errorscode.CODE_ENTITY_NOT_FOUND)
			return
		}
		response.InternalError(c.logger, ginCtx, err, ginCtx.GetString(string(errorscode.CtxErrorCodeKey)))
		return
	}

	ginCtx.JSON(http.StatusOK, resp)
}

func (c *controller) UpdateUser(ginCtx *gin.Context) {
	var request dto.UpdateUserRequest
	ctx := ginCtx.Request.Context()

	err := ginCtx.BindJSON(&request)
	if err != nil {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_REQUEST)
		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_REQUEST)
		return
	}

	err = c.usecases.Update(&ctx, &request)
	if err != nil {
		response.InternalError(c.logger, ginCtx, err, ginCtx.GetString(string(errorscode.CtxErrorCodeKey)))
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "user updated"})
}

func (c *controller) UpdateEmail(ginCtx *gin.Context) {
	var request dto.UpdateEmailRequest
	ctx := ginCtx.Request.Context()

	userUUID := ginCtx.Param(UUID)

	err := uuid.ValidateWithPrefix(userUUID)
	if err != nil {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_UUID)
		return
	}

	err = ginCtx.BindJSON(&request)
	if err != nil {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_REQUEST)
		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_REQUEST)
		return
	}

	err = c.usecases.UpdateEmail(&ctx, userUUID, request.Email)
	if err != nil {
		response.InternalError(c.logger, ginCtx, err, ginCtx.GetString(string(errorscode.CtxErrorCodeKey)))
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "email updated"})
}

func (c *controller) UpdatePassword(ginCtx *gin.Context) {
	var request dto.UpdatePasswordRequest
	ctx := ginCtx.Request.Context()

	userUUID := ginCtx.Param(UUID)

	err := uuid.ValidateWithPrefix(userUUID)
	if err != nil {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_UUID)
		return
	}

	err = ginCtx.BindJSON(&request)
	if err != nil {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_REQUEST)
		return
	}

	err = c.validate.Struct(request)
	if err != nil {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_REQUEST)
		return
	}

	err = c.usecases.UpdatePassword(&ctx, userUUID, request.Password)

	if err != nil {
		response.InternalError(c.logger, ginCtx, err, ginCtx.GetString(string(errorscode.CtxErrorCodeKey)))
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

func (c *controller) DeleteUser(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()

	userUUID := ginCtx.Param(UUID)

	err := uuid.ValidateWithPrefix(userUUID)
	if err != nil {
		response.BadRequest(c.logger, ginCtx, errorscode.CODE_INVALID_UUID)
		return
	}

	err = c.usecases.Delete(&ctx, userUUID)
	if err != nil {
		response.InternalError(c.logger, ginCtx, err, ginCtx.GetString(string(errorscode.CtxErrorCodeKey)))
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func (c *controller) GetAllUsers(ginCtx *gin.Context) {
	ctx := ginCtx.Request.Context()

	resp, err := c.usecases.GetAllUsers(&ctx)
	if err != nil {
		response.InternalError(c.logger, ginCtx, err, ginCtx.GetString(string(errorscode.CtxErrorCodeKey)))
		return
	}

	ginCtx.JSON(http.StatusOK, resp)
}
