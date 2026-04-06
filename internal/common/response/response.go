package response

import (
	"errors"
	"net/http"

	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/middlewares"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type errorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	TraceID string `json:"trace_id"`
}

func Error(logger *zap.SugaredLogger, c *gin.Context, err error) {
	traceID := middlewares.GetTraceID(c)

	var apiErr *errorsutil.AppError
	if ok := errors.As(err, &apiErr); !ok {
		logger.Errorw("unhandled error", "trace_id", traceID, "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse{
			Message: "internal server error",
			Code:    errorsutil.ErrorUnknownError,
			TraceID: traceID,
		})

		return
	}

	if apiErr.Code == 0 {
		apiErr.Code = errorsutil.ErrorUnknownError
	}

	httpCode := apiErr.MapErrorCodeToHTTPCode()
	logger.Errorw("request error", "trace_id", traceID, "code", apiErr.Code, "http_status", httpCode, "message", apiErr.Message, "error", apiErr.Err)

	c.JSON(httpCode, errorResponse{
		Message: apiErr.Message,
		Code:    apiErr.Code,
		TraceID: traceID,
	})
}

func BadRequest(logger *zap.SugaredLogger, c *gin.Context, appErrCode int, message string) {
	traceID := middlewares.GetTraceID(c)
	logger.Warnw("bad request", "trace_id", traceID, "code", appErrCode, "message", message)

	c.JSON(http.StatusBadRequest, errorResponse{
		Message: message,
		Code:    appErrCode,
		TraceID: traceID,
	})
}
