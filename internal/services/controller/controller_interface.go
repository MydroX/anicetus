package controller

import "github.com/gin-gonic/gin"

type ControllerInterface interface {
	RegisterService(ginCtx *gin.Context)
	RevokeService(ginCtx *gin.Context)
	GetAllServices(ginCtx *gin.Context)
	GetUserServices(ginCtx *gin.Context)
	AssignServiceToUser(ginCtx *gin.Context)
}
