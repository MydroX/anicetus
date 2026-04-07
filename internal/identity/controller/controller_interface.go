package controller

import "github.com/gin-gonic/gin"

type ControllerInterface interface {
	CreateUser(ginCtx *gin.Context)
	GetUser(ginCtx *gin.Context)
	UpdateUser(ginCtx *gin.Context)
	UpdateEmail(ginCtx *gin.Context)
	UpdatePassword(ginCtx *gin.Context)
	DeleteUser(ginCtx *gin.Context)
	GetAllUsers(ginCtx *gin.Context)
}
