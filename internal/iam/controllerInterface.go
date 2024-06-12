package iam

import "github.com/gin-gonic/gin"

//go:generate mockgen -destination=../mocks/mock_controller.go -package=mocks MydroX/project-v/internal/iam/controller ControllerInterface

type ControllerInterface interface {
	CreateUser(c *gin.Context)
	GetUser(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
	AuthenticateUser(c *gin.Context)
}
