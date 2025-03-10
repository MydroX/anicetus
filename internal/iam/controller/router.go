package iam

import "github.com/gin-gonic/gin"

func Router(e *gin.RouterGroup, c ControllerInterface) {
	// Public routes
	e.POST("/login", c.Login)

	// Logged in routes
	e.GET("/refresh", c.RefreshToken)
}
