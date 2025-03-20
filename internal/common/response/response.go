package response

import (
	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errors"
	loggerpkg "MydroX/anicetus/pkg/logger"
	"fmt"
	"net/http"
	"os"
)

const (
	dev     = "dev"
	staging = "staging"
	prod    = "prod"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	TraceId string `json:"trace_id"`
}

type ErrorOption func(*errorOptions)

type errorOptions struct {
	logMessage    string
	clientMessage string
}

// WithLogMessage sets a custom log message
func WithLogMessage(msg string, args ...any) ErrorOption {
	return func(o *errorOptions) {
		if len(args) > 0 {
			o.logMessage = fmt.Sprintf(msg, args...)
		} else {
			o.logMessage = msg
		}
	}
}

// WithClientMessage sets a custom client-facing message
func WithClientMessage(msg string) ErrorOption {
	return func(o *errorOptions) {
		o.clientMessage = msg
	}
}

// WithDebugLog sets the log message to be the error message only when debug is enabled in environment
func WithDebugLog(msg string, args ...any) ErrorOption {
	return func(o *errorOptions) {
		if len(args) > 0 {
			o.logMessage = fmt.Sprintf(msg, args...)
		} else {
			o.logMessage = msg
		}
	}
}

// Error sends an error response with the given HTTP status code
func Error(logger *loggerpkg.Logger, ctx *context.AppContext, httpCode int, apiErrorCode string, opts ...ErrorOption) {
	// Default options
	options := &errorOptions{
		logMessage:    "Error occurred",
		clientMessage: "An error occurred",
	}

	// Apply provided options
	for _, opt := range opts {
		opt(options)
	}

	if apiErrorCode == "" {
		apiErrorCode = errors.ERROR_UNKNOWN_ERROR
	}

	// Get trace ID
	traceID := ctx.EnsureTraceID()

	// Log debug
	if os.Getenv("env") == dev || os.Getenv("env") == staging {
		logger.Zap.Debug(fmt.Sprintf("[%d] | [%s] | %s | %s",
			httpCode, apiErrorCode, traceID, options.logMessage))
	}

	// Log the error
	logger.Zap.Error(fmt.Sprintf("[%d] | [%s] | %s | %s",
		httpCode, apiErrorCode, traceID, options.logMessage))

	ctx.GinContext().JSON(httpCode, ErrorResponse{
		Message: options.clientMessage,
		Code:    apiErrorCode,
		TraceId: traceID,
	})
}

// InternalError sends a 500 Internal Server Error response
func InternalError(logger *loggerpkg.Logger, ctx *context.AppContext, err *errors.Err, opts ...ErrorOption) {
	defaultOpts := []ErrorOption{
		WithLogMessage(err.Err.Error()),
		WithClientMessage("Internal server error"),
	}
	Error(logger, ctx, http.StatusInternalServerError, err.Code, append(defaultOpts, opts...)...)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(logger *loggerpkg.Logger, ctx *context.AppContext, err *errors.Err, opts ...ErrorOption) {
	defaultOpts := []ErrorOption{
		WithLogMessage(err.Message),
		WithClientMessage("Invalid request"),
	}
	Error(logger, ctx, http.StatusBadRequest, err.Code, append(defaultOpts, opts...)...)
}

// Conflict sends a 409 Conflict response
func Conflict(logger *loggerpkg.Logger, ctx *context.AppContext, err *errors.Err, opts ...ErrorOption) {
	defaultOpts := []ErrorOption{
		WithLogMessage(err.Message),
		WithClientMessage("Resource conflict"),
	}
	Error(logger, ctx, http.StatusConflict, err.Code, append(defaultOpts, opts...)...)
}

// NotFound sends a 404 Not Found response
func NotFound(logger *loggerpkg.Logger, ctx *context.AppContext, err *errors.Err, opts ...ErrorOption) {
	defaultOpts := []ErrorOption{
		WithLogMessage(err.Message),
		WithClientMessage("Entity not found"),
	}
	Error(logger, ctx, http.StatusNotFound, err.Code, append(defaultOpts, opts...)...)
}
