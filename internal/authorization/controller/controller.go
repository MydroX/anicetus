package controller

import (
	"net/http"

	"MydroX/anicetus/internal/authorization/dto"
	"MydroX/anicetus/internal/authorization/usecases"
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
	usecases usecases.AuthorizationUsecases
}

func New(l *zap.SugaredLogger, u usecases.AuthorizationUsecases) ControllerInterface {
	v := validator.New()

	return &controller{
		validate: v,
		logger:   l,
		usecases: u,
	}
}

// Roles

func (c *controller) CreateRole(ginCtx *gin.Context) {
	var request dto.CreateRoleRequest

	if err := ginCtx.BindJSON(&request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorFailToBind, errs.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, errs.MessageInvalidInput)

		return
	}

	if err := c.usecases.CreateRole(ginCtx.Request.Context(), &request); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusCreated, gin.H{"message": "role created"})
}

func (c *controller) GetAllRoles(ginCtx *gin.Context) {
	roles, err := c.usecases.GetAllRoles(ginCtx.Request.Context())
	if err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"roles": roles})
}

func (c *controller) UpdateRole(ginCtx *gin.Context) {
	roleUUID := ginCtx.Param("uuid")

	if _, err := uuid.Parse(roleUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	var request dto.UpdateRoleRequest

	if err := ginCtx.BindJSON(&request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorFailToBind, errs.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, errs.MessageInvalidInput)

		return
	}

	if err := c.usecases.UpdateRole(ginCtx.Request.Context(), roleUUID, &request); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

func (c *controller) DeleteRole(ginCtx *gin.Context) {
	roleUUID := ginCtx.Param("uuid")

	if _, err := uuid.Parse(roleUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	if err := c.usecases.DeleteRole(ginCtx.Request.Context(), roleUUID); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "role deleted"})
}

// Permissions

func (c *controller) CreatePermission(ginCtx *gin.Context) {
	var request dto.CreatePermissionRequest

	if err := ginCtx.BindJSON(&request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorFailToBind, errs.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, errs.MessageInvalidInput)

		return
	}

	if err := c.usecases.CreatePermission(ginCtx.Request.Context(), &request); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusCreated, gin.H{"message": "permission created"})
}

func (c *controller) GetAllPermissions(ginCtx *gin.Context) {
	permissions, err := c.usecases.GetAllPermissions(ginCtx.Request.Context())
	if err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// Role-Permission assignments

//nolint:dupl // Similar structure to AssignRoleToUser but different request/response types
func (c *controller) AssignPermissionToRole(ginCtx *gin.Context) {
	roleUUID := ginCtx.Param("uuid")

	if _, err := uuid.Parse(roleUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	var request dto.AssignPermissionToRoleRequest

	if err := ginCtx.BindJSON(&request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorFailToBind, errs.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, errs.MessageInvalidInput)

		return
	}

	if err := c.usecases.AssignPermissionToRole(ginCtx.Request.Context(), roleUUID, request.PermissionUUID); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "permission assigned to role"})
}

func (c *controller) RemovePermissionFromRole(ginCtx *gin.Context) {
	roleUUID := ginCtx.Param("uuid")
	permUUID := ginCtx.Param("perm_uuid")

	if _, err := uuid.Parse(roleUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	if _, err := uuid.Parse(permUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	if err := c.usecases.RemovePermissionFromRole(ginCtx.Request.Context(), roleUUID, permUUID); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "permission removed from role"})
}

// User-Role assignments

//nolint:dupl // Similar structure to AssignPermissionToRole but different request/response types
func (c *controller) AssignRoleToUser(ginCtx *gin.Context) {
	userUUID := ginCtx.Param("uuid")

	if _, err := uuid.Parse(userUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	var request dto.AssignRoleToUserRequest

	if err := ginCtx.BindJSON(&request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorFailToBind, errs.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, errs.MessageInvalidInput)

		return
	}

	if err := c.usecases.AssignRoleToUser(ginCtx.Request.Context(), userUUID, request.RoleUUID); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "role assigned to user"})
}

func (c *controller) RemoveRoleFromUser(ginCtx *gin.Context) {
	userUUID := ginCtx.Param("uuid")
	roleUUID := ginCtx.Param("role_uuid")

	if _, err := uuid.Parse(userUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	if _, err := uuid.Parse(roleUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	if err := c.usecases.RemoveRoleFromUser(ginCtx.Request.Context(), userUUID, roleUUID); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "role removed from user"})
}

func (c *controller) GetUserRoles(ginCtx *gin.Context) {
	userUUID := ginCtx.Param("uuid")

	if _, err := uuid.Parse(userUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	roles, err := c.usecases.GetUserRoles(ginCtx.Request.Context(), userUUID)
	if err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"roles": roles})
}

func (c *controller) GetUserPermissions(ginCtx *gin.Context) {
	userUUID := ginCtx.Param("uuid")

	if _, err := uuid.Parse(userUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	permissions, err := c.usecases.GetUserPermissions(ginCtx.Request.Context(), userUUID)
	if err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, dto.UserPermissionsResponse{Permissions: permissions})
}
