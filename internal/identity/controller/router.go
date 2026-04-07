package controller

import "github.com/gin-gonic/gin"

func PublicRouter(v1 *gin.RouterGroup, c ControllerInterface) {
	v1.POST("/register", c.CreateUser)
}

func AuthenticatedRouter(authenticated *gin.RouterGroup, c ControllerInterface) {
	users := authenticated.Group("/users")

	// Logged in routes
	users.PUT("/:uuid", c.UpdateUser)
	users.PATCH("/:uuid/email", c.UpdateEmail)
	users.PATCH("/:uuid/password", c.UpdatePassword)

	// Admin routes
	users.GET("/:uuid", c.GetUser)
	users.DELETE("/:uuid", c.DeleteUser)
	users.GET("/", c.GetAllUsers)
}
