package iam

import "github.com/gin-gonic/gin"

type ControllerInterface interface {
	Login(ginCtx *gin.Context)
	RefreshToken(ginCtx *gin.Context)
}
