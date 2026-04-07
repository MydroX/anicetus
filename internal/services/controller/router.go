package controller

import "github.com/gin-gonic/gin"

func Router(e *gin.RouterGroup, c ControllerInterface) {
	services := e.Group("/services")
	services.POST("", c.RegisterService)
	services.DELETE("/:service", c.RevokeService)
	services.GET("", c.GetAllServices)
	services.GET("/users/:uuid", c.GetUserServices)
	services.POST("/users/:uuid", c.AssignServiceToUser)
}
