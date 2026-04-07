package controller

import "github.com/gin-gonic/gin"

type ControllerInterface interface {
	Login(ginCtx *gin.Context)
	Logout(ginCtx *gin.Context)
	RefreshToken(ginCtx *gin.Context)
}
