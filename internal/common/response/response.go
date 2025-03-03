package response

import (
	"MydroX/project-v/internal/common/errorscode"
	loggerpkg "MydroX/project-v/pkg/logger"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// logAndError is a function to handle logAndError response
// The function will log the logAndError message and return the logAndError message to the client.
// In the case of debug mode, the logAndError message will be shown, otherwise, a generic logAndError message will be shown.
func logAndError(logger *loggerpkg.Logger, ctx *gin.Context, httpCode int, errorMessage, apiErrorCode, messageAPI string) {
	var errorCode string

	if apiErrorCode == "" {
		errorCode = errorscode.CODE_UNKNOWN_ERROR
	} else {
		errorCode = apiErrorCode
	}

	logger.Zap.Error(fmt.Sprintf("[%d] | [%s] | %s", httpCode, errorCode, errorMessage))
	if logger.Debug {
		ctx.JSON(httpCode, gin.H{"error": errorMessage})
		return
	}

	ctx.JSON(httpCode, ErrorResponse{
		Message: messageAPI,
		Code:    errorCode,
	})
}

// InternalError is a function to handle error response for internal server error
func InternalError(logger *loggerpkg.Logger, ctx *gin.Context, err error, apiErrorCode string) {
	logAndError(logger, ctx, http.StatusInternalServerError, err.Error(), apiErrorCode, "internal error")
}

// BadRequest is a function to handle error response for invalid request
func BadRequest(logger *loggerpkg.Logger, ctx *gin.Context, apiErrorCode string) {
	logAndError(logger, ctx, http.StatusBadRequest, "invalid request", apiErrorCode, "invalid request")
}

func BadRequestWithMessage(logger *loggerpkg.Logger, ctx *gin.Context, apiErrorCode, message string) {
	logAndError(logger, ctx, http.StatusBadRequest, "invalid request", apiErrorCode, message)
}

func Conflict(logger *loggerpkg.Logger, ctx *gin.Context, apiErrorCode string) {
	logAndError(logger, ctx, http.StatusConflict, "conflict", apiErrorCode, "conflict")
}

// NotFound is a function to handle error response for not found entity
func NotFound(logger *loggerpkg.Logger, ctx *gin.Context, apiErrorCode string) {
	logAndError(logger, ctx, http.StatusNotFound, "not found", apiErrorCode, "entity not found")
}
