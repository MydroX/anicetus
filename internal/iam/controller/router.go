package iam

import "github.com/gin-gonic/gin"

func Router(e *gin.RouterGroup, c ControllerInterface) {
	// Public routes
	e.POST("/login", c.Login)

	// Logged in routes
	e.GET("/refresh", c.RefreshToken)

	// Audience management
	audiences := e.Group("/audiences")
	audiences.POST("", c.RegisterAudience)
	audiences.DELETE("/:audience", c.RevokeAudience)
	audiences.GET("", c.GetAllAudiences)
	audiences.GET("/users/:uuid", c.GetUserAudiences)
	audiences.POST("/users/:uuid", c.AssignAudienceToUser)
}
