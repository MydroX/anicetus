package response

import (
	"errors"
	"fmt"
	"net/http"

	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errorsutil"

	"go.uber.org/zap"
)

type errorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	TraceID string `json:"trace_id"`
}

// Error sends an error response, extracting the HTTP status from the AppError code.
func Error(logger *zap.SugaredLogger, ctx *context.AppContext, err error) {
	traceID := ctx.EnsureTraceID()

	var apiErr *errorsutil.AppError
	if ok := errors.As(err, &apiErr); !ok {
		logger.Errorw("unhandled error", "trace_id", traceID, "error", err)
		ctx.GinContext().JSON(http.StatusInternalServerError, errorResponse{
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

	ctx.GinContext().JSON(httpCode, errorResponse{
		Message: apiErr.Message,
		Code:    apiErr.Code,
		TraceID: traceID,
	})
}

// BadRequest sends a 400 Bad Request response with the given error code and message.
func BadRequest(logger *zap.SugaredLogger, ctx *context.AppContext, appErrCode int, message string) {
	traceID := ctx.EnsureTraceID()
	logger.Warnw("bad request", "trace_id", traceID, "code", appErrCode, "message", message)

	ctx.GinContext().JSON(http.StatusBadRequest, errorResponse{
		Message: message,
		Code:    appErrCode,
		TraceID: traceID,
	})
}

// Success sends a JSON response with the given status code and data.
func Success(ctx *context.AppContext, statusCode int, data any) {
	ctx.GinContext().JSON(statusCode, data)
}

// Created sends a 201 response with the given data.
func Created(ctx *context.AppContext, data any) {
	ctx.GinContext().JSON(http.StatusCreated, data)
}

// NoContent sends a 204 response.
func NoContent(ctx *context.AppContext) {
	ctx.GinContext().Status(http.StatusNoContent)
}

// NotFound sends a 404 response.
func NotFound(logger *zap.SugaredLogger, ctx *context.AppContext, message string) {
	traceID := ctx.EnsureTraceID()
	logger.Warnw("not found", "trace_id", traceID, "message", message)

	ctx.GinContext().JSON(http.StatusNotFound, errorResponse{
		Message: message,
		Code:    errorsutil.ErrorNotFound,
		TraceID: traceID,
	})
}

// Unauthorized sends a 401 response.
func Unauthorized(logger *zap.SugaredLogger, ctx *context.AppContext, message string) {
	traceID := ctx.EnsureTraceID()
	logger.Warnw("unauthorized", "trace_id", traceID, "message", message)

	ctx.GinContext().JSON(http.StatusUnauthorized, errorResponse{
		Message: message,
		Code:    errorsutil.ErrorUnauthorized,
		TraceID: traceID,
	})
}

// InternalError sends a 500 response with a generic message.
func InternalError(logger *zap.SugaredLogger, ctx *context.AppContext, err error) {
	traceID := ctx.EnsureTraceID()
	logger.Errorw("internal error", "trace_id", traceID, "error", fmt.Sprintf("%v", err))

	ctx.GinContext().JSON(http.StatusInternalServerError, errorResponse{
		Message: "internal server error",
		Code:    errorsutil.ErrorInternal,
		TraceID: traceID,
	})
}
