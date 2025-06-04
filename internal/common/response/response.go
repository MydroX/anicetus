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

type ErrorOptions func(*errorOptions)

type errorOptions struct {
	clientMessage string
}

// WithClientMessage sets a custom client-facing message
func WithClientMessage(msg string) ErrorOptions {
	return func(o *errorOptions) {
		o.clientMessage = msg
	}
}

// handleError sends an error response with the given HTTP status code
func handleError(logger *zap.SugaredLogger, ctx *context.AppContext, appErr *errorsutil.AppError, options *errorOptions) {
	if appErr.Severity == "" {
		logger.Warn(fmt.Sprintf("Severity is not set for request %s", ctx.EnsureTraceID()))

		appErr.Severity = errorsutil.SeverityError
	}

	// Get trace ID
	traceID := ctx.EnsureTraceID()

	// Get HTTP status code
	httpCode := appErr.MapErrorCodeToHTTPCode()

	// Log error
	logger.Error(fmt.Sprintf("%s | %d | %d | %s | %s \n%v", appErr.Severity, httpCode, appErr.Code, traceID, appErr.Message, appErr.Err))

	ctx.GinContext().JSON(httpCode, errorResponse{
		Message: options.clientMessage,
		Code:    appErr.Code,
		TraceID: traceID,
	})
}

// applyOptions applies the default options and any provided options
func applyOptions(opts ...ErrorOptions) *errorOptions {
	options := &errorOptions{
		clientMessage: "An error occurred",
	}

	// Apply provided options
	for _, opt := range opts {
		opt(options)
	}

	return options
}

// Error is the generic error handler
func Error(logger *zap.SugaredLogger, ctx *context.AppContext, err error, opts ...ErrorOptions) {
	var apiErr *errorsutil.AppError
	if ok := errors.As(err, &apiErr); !ok {
		logger.Error(fmt.Sprintf("CRITICAL | [%s] | %s | %s ", ctx.EnsureTraceID(), "Something went wrong while handling error", err))
		ctx.GinContext().JSON(http.StatusInternalServerError, errorResponse{
			Message: "Internal server error, server has not been able to handle the error properly",
			Code:    errorsutil.ErrorUnknownError,
			TraceID: ctx.EnsureTraceID(),
		})

		return
	}

	// Apply options
	options := applyOptions(opts...)

	if apiErr.Code == 0 {
		apiErr.Code = errorsutil.ErrorUnknownError
	}

	handleError(logger, ctx, apiErr, options)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(logger *zap.SugaredLogger, ctx *context.AppContext, appErrCode int, message string) {
	appErr := &errorsutil.AppError{
		Code:     appErrCode,
		Message:  "Bad request: " + message,
		Severity: errorsutil.SeverityError,
	}

	opts := []ErrorOptions{
		WithClientMessage(message),
	}

	options := applyOptions(opts...)
	handleError(logger, ctx, appErr, options)
}
