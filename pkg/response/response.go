package response

import (
	"MydroX/project-v/pkg/logger"
	"fmt"

	"github.com/gin-gonic/gin"
)

// logAndError is a function to handle logAndError response
// The function will log the logAndError message and return the logAndError message to the client.
// In the case of debug mode, the logAndError message will be shown, otherwise, a generic logAndError message will be shown.
func logAndError(logger *logger.Logger, ctx *gin.Context, code int, message string) {
	logger.Zap.Error(fmt.Sprintf("[%d] %s", code, message))
	if logger.Debug {
		ctx.JSON(code, gin.H{"error": message})
		return
	}
	ctx.JSON(code, gin.H{"error": "an error occurred. please try again later or contact support"})
}

// InternalError is a function to handle error response for internal server error
func InternalError(logger *logger.Logger, ctx *gin.Context, err error) {
	logAndError(logger, ctx, 500, err.Error())
}

// InvalidRequest is a function to handle error response for invalid request
func InvalidRequest(logger *logger.Logger, ctx *gin.Context) {
	logAndError(logger, ctx, 400, "invalid request")
}

// NotFound is a function to handle error response for not found entity
func NotFound(logger *logger.Logger, ctx *gin.Context) {
	logAndError(logger, ctx, 404, "not found")
}
