package controller

import "github.com/gin-gonic/gin"

func PublicRouter(v1 *gin.RouterGroup, c ControllerInterface) {
	v1.POST("/login", c.Login)
	v1.POST("/refresh", c.RefreshToken)
}

func AuthenticatedRouter(v1 *gin.RouterGroup, c ControllerInterface) {
	v1.POST("/logout", c.Logout)
}
