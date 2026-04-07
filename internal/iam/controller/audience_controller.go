package iam

import (
	"net/http"

	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/httpresponse"
	"MydroX/anicetus/internal/iam/dto"
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
)

func (c *controller) RegisterAudience(ginCtx *gin.Context) {
	var request dto.RegisterAudienceRequest


	if err := ginCtx.BindJSON(&request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorFailToBind, errs.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, errs.MessageInvalidInput)

		return
	}

	if err := c.usecases.RegisterAudience(ginCtx.Request.Context(), &request); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusCreated, gin.H{"message": "audience registered"})
}

func (c *controller) RevokeAudience(ginCtx *gin.Context) {

	audience := ginCtx.Param("audience")
	if audience == "" {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, "audience is required")

		return
	}

	if err := c.usecases.RevokeAudience(ginCtx.Request.Context(), audience); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "audience revoked"})
}

func (c *controller) GetAllAudiences(ginCtx *gin.Context) {

	audiences, err := c.usecases.GetAllAudiences(ginCtx.Request.Context())
	if err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, dto.AudienceListResponse{Audiences: audiences})
}

func (c *controller) GetUserAudiences(ginCtx *gin.Context) {

	userUUID := ginCtx.Param("uuid")

	if _, err := uuid.Parse(userUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	audiences, err := c.usecases.GetUserAudiences(ginCtx.Request.Context(), userUUID)
	if err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, dto.AudienceListResponse{Audiences: audiences})
}

func (c *controller) AssignAudienceToUser(ginCtx *gin.Context) {
	var request dto.AssignAudienceRequest


	userUUID := ginCtx.Param("uuid")

	if _, err := uuid.Parse(userUUID); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidUUID, errs.MessageInvalidUUID)

		return
	}

	if err := ginCtx.BindJSON(&request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorFailToBind, errs.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		httpresponse.BadRequest(c.logger, ginCtx, errs.ErrorInvalidInput, errs.MessageInvalidInput)

		return
	}

	if err := c.usecases.AssignAudienceToUser(ginCtx.Request.Context(), userUUID, &request); err != nil {
		httpresponse.Error(c.logger, ginCtx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "audience assigned to user"})
}
