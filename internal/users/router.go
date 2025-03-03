package users

import "github.com/gin-gonic/gin"

func Router(v1 *gin.RouterGroup, c *Controller) {
	users := v1.Group("/users")

	// Public routes
	users.POST("/register", c.CreateUser)
	users.POST("/login", c.Login)

	// Logged in routes
	users.PUT("/:uuid", c.UpdateUser)
	users.PATCH("/:uuid/email", c.UpdateEmail)
	users.PATCH("/:uuid/password", c.UpdatePassword)

	// Admin routes
	users.GET("/:uuid", c.GetUser)
	users.DELETE("/:uuid", c.DeleteUser)
	users.GET("/", c.GetAllUsers)
}
