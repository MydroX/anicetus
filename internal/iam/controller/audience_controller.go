package iam

import (
	"net/http"

	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/common/response"
	"MydroX/anicetus/internal/iam/dto"
	"MydroX/anicetus/pkg/uuid"
	"github.com/gin-gonic/gin"
)

func (c *controller) RegisterAudience(ginCtx *gin.Context) {
	var request dto.RegisterAudienceRequest

	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	if err := ginCtx.BindJSON(&request); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorFailToBind, errorsutil.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidInput, errorsutil.MessageInvalidInput)

		return
	}

	if err := c.usecases.RegisterAudience(ctx, &request); err != nil {
		response.Error(c.logger, ctx, err)

		return
	}

	ginCtx.JSON(http.StatusCreated, gin.H{"message": "audience registered"})
}

func (c *controller) RevokeAudience(ginCtx *gin.Context) {
	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	audience := ginCtx.Param("audience")
	if audience == "" {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidInput, "audience is required")

		return
	}

	if err := c.usecases.RevokeAudience(ctx, audience); err != nil {
		response.Error(c.logger, ctx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "audience revoked"})
}

func (c *controller) GetAllAudiences(ginCtx *gin.Context) {
	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	audiences, err := c.usecases.GetAllAudiences(ctx)
	if err != nil {
		response.Error(c.logger, ctx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, dto.AudienceListResponse{Audiences: audiences})
}

func (c *controller) GetUserAudiences(ginCtx *gin.Context) {
	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	userUUID := ginCtx.Param("uuid")

	if _, err := uuid.ValidateWithPrefix(userUUID); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidUUID, errorsutil.MessageInvalidUUID)

		return
	}

	audiences, err := c.usecases.GetUserAudiences(ctx, userUUID)
	if err != nil {
		response.Error(c.logger, ctx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, dto.AudienceListResponse{Audiences: audiences})
}

func (c *controller) AssignAudienceToUser(ginCtx *gin.Context) {
	var request dto.AssignAudienceRequest

	ctx := context.NewAppContext(ginCtx)
	ctx.EnsureTraceID()

	userUUID := ginCtx.Param("uuid")

	if _, err := uuid.ValidateWithPrefix(userUUID); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidUUID, errorsutil.MessageInvalidUUID)

		return
	}

	if err := ginCtx.BindJSON(&request); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorFailToBind, errorsutil.MessageFailToBind)

		return
	}

	if err := c.validate.Struct(request); err != nil {
		response.BadRequest(c.logger, ctx, errorsutil.ErrorInvalidInput, errorsutil.MessageInvalidInput)

		return
	}

	if err := c.usecases.AssignAudienceToUser(ctx, userUUID, &request); err != nil {
		response.Error(c.logger, ctx, err)

		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "audience assigned to user"})
}
