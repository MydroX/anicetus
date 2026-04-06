package iam

import "github.com/gin-gonic/gin"

type ControllerInterface interface {
	Login(ginCtx *gin.Context)
	RefreshToken(ginCtx *gin.Context)

	// Audience management
	RegisterAudience(ginCtx *gin.Context)
	RevokeAudience(ginCtx *gin.Context)
	GetAllAudiences(ginCtx *gin.Context)
	GetUserAudiences(ginCtx *gin.Context)
	AssignAudienceToUser(ginCtx *gin.Context)
}
