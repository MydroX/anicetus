package controller

import "github.com/gin-gonic/gin"

func Router(e *gin.RouterGroup, c ControllerInterface) {
	// Roles
	roles := e.Group("/roles")
	roles.POST("", c.CreateRole)
	roles.GET("", c.GetAllRoles)
	roles.PUT("/:uuid", c.UpdateRole)
	roles.DELETE("/:uuid", c.DeleteRole)
	roles.POST("/:uuid/permissions", c.AssignPermissionToRole)
	roles.DELETE("/:uuid/permissions/:perm_uuid", c.RemovePermissionFromRole)

	// Permissions
	permissions := e.Group("/permissions")
	permissions.POST("", c.CreatePermission)
	permissions.GET("", c.GetAllPermissions)

	// User assignments
	e.POST("/users/:uuid/roles", c.AssignRoleToUser)
	e.DELETE("/users/:uuid/roles/:role_uuid", c.RemoveRoleFromUser)
	e.GET("/users/:uuid/roles", c.GetUserRoles)
	e.GET("/users/:uuid/permissions", c.GetUserPermissions)
}
