package controller

import "github.com/gin-gonic/gin"

type ControllerInterface interface {
	CreateRole(ginCtx *gin.Context)
	GetAllRoles(ginCtx *gin.Context)
	UpdateRole(ginCtx *gin.Context)
	DeleteRole(ginCtx *gin.Context)

	CreatePermission(ginCtx *gin.Context)
	GetAllPermissions(ginCtx *gin.Context)

	AssignPermissionToRole(ginCtx *gin.Context)
	RemovePermissionFromRole(ginCtx *gin.Context)

	AssignRoleToUser(ginCtx *gin.Context)
	RemoveRoleFromUser(ginCtx *gin.Context)
	GetUserRoles(ginCtx *gin.Context)
	GetUserPermissions(ginCtx *gin.Context)
}
