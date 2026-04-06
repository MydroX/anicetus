package iam

import (
	"net/http"

	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/common/response"
	"MydroX/anicetus/internal/iam/dto"
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
)

func (c *controller) RegisterAudience(ginCtx *gin.Context) {
	var request dto.RegisterAudienceRequest


	if err := ginCtx.BindJSON(&request); err != nil {
		response.BadRequest(c.logger, ginCtx, errorsutil.ErrorFailToBind, errorsutil.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		response.BadRequest(c.logger, ginCtx, errorsutil.ErrorInvalidInput, errorsutil.MessageInvalidInput)

		return
	}

	if err := c.usecases.RegisterAudience(ginCtx.Request.Context(), &request); err != nil {
		response.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusCreated, gin.H{"message": "audience registered"})
}

func (c *controller) RevokeAudience(ginCtx *gin.Context) {

	audience := ginCtx.Param("audience")
	if audience == "" {
		response.BadRequest(c.logger, ginCtx, errorsutil.ErrorInvalidInput, "audience is required")

		return
	}

	if err := c.usecases.RevokeAudience(ginCtx.Request.Context(), audience); err != nil {
		response.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "audience revoked"})
}

func (c *controller) GetAllAudiences(ginCtx *gin.Context) {

	audiences, err := c.usecases.GetAllAudiences(ginCtx.Request.Context())
	if err != nil {
		response.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, dto.AudienceListResponse{Audiences: audiences})
}

func (c *controller) GetUserAudiences(ginCtx *gin.Context) {

	userUUID := ginCtx.Param("uuid")

	if _, err := uuid.Parse(userUUID); err != nil {
		response.BadRequest(c.logger, ginCtx, errorsutil.ErrorInvalidUUID, errorsutil.MessageInvalidUUID)

		return
	}

	audiences, err := c.usecases.GetUserAudiences(ginCtx.Request.Context(), userUUID)
	if err != nil {
		response.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, dto.AudienceListResponse{Audiences: audiences})
}

func (c *controller) AssignAudienceToUser(ginCtx *gin.Context) {
	var request dto.AssignAudienceRequest


	userUUID := ginCtx.Param("uuid")

	if _, err := uuid.Parse(userUUID); err != nil {
		response.BadRequest(c.logger, ginCtx, errorsutil.ErrorInvalidUUID, errorsutil.MessageInvalidUUID)

		return
	}

	if err := ginCtx.BindJSON(&request); err != nil {
		response.BadRequest(c.logger, ginCtx, errorsutil.ErrorFailToBind, errorsutil.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		response.BadRequest(c.logger, ginCtx, errorsutil.ErrorInvalidInput, errorsutil.MessageInvalidInput)

		return
	}

	if err := c.usecases.AssignAudienceToUser(ginCtx.Request.Context(), userUUID, &request); err != nil {
		response.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "audience assigned to user"})
}
