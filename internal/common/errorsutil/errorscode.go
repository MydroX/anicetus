package errorsutil

import "net/http"

// Error codes with numeric values
// Format: CCXXXX where:
// - CC: Category code (2 digits)
// - XXXX: Error code (4 digits)
const (

	// Common errors (10xxx)
	ERROR_FAIL_TO_BIND     = 10001 // Failed to bind request
	ERROR_INVALID_INPUT    = 10002 // Invalid input
	ERROR_INVALID_UUID     = 10003 // Invalid UUID
	ERROR_NOT_FOUND        = 10004 // Not found
	ERROR_DUPLICATE_ENTITY = 10005 // Duplicate entity
	ERROR_UNAUTHORIZED     = 10006 // Unauthorized
	ERROR_TOO_MANY_REQUEST = 10007 // Too many requests

	ERROR_INTERNAL      = 10008 // Internal server error
	ERROR_UNKNOWN_ERROR = 10999 // Unknown error

	// User errors (11xxx)
	ERROR_INVALID_USERNAME        = 11001 // Username does not meet the requirements
	ERROR_FAILED_TO_HASH_PASSWORD = 11002 // Failed to hash password
	ERROR_INVALID_PASSWORD        = 11003 // Password does not meet the requirements

	// IAM errors (12xxx)
	ERROR_HASH_TOKEN   = 12001 // Failed to hash token
	ERROR_CREATE_TOKEN = 12002 // Failed to create token

	// Database errors (99xxx)
	ERROR_UNIQUE_VIOLATION      = 99001 // Unique constraint violation
	ERROR_FOREIGN_KEY_VIOLATION = 99002 // Foreign key constraint violation
	ERROR_CONSTRAINT_VIOLATION  = 99003 // Check constraint violation
	ERROR_NOT_NULL_VIOLATION    = 99004 // Not null constraint violation
	ERROR_DATABASE_UNAVAILABLE  = 99005 // Database unavailable
	ERROR_UNKNOWN_ERROR_DB      = 99999 // Unknown database error
)

const (
	MessageFailToBind   = "Failed to bind request. Please check your request and try again"
	MessageInvalidInput = "Invalid input. Please check your request and try again"
	MessageInvalidUUID  = "Invalid UUID"
)

var errorCodeMap = map[int]int{
	// 400 Bad Request
	ERROR_FAIL_TO_BIND:         http.StatusBadRequest,
	ERROR_INVALID_INPUT:        http.StatusBadRequest,
	ERROR_INVALID_UUID:         http.StatusBadRequest,
	ERROR_INVALID_USERNAME:     http.StatusBadRequest,
	ERROR_INVALID_PASSWORD:     http.StatusBadRequest,
	ERROR_CONSTRAINT_VIOLATION: http.StatusBadRequest,
	ERROR_NOT_NULL_VIOLATION:   http.StatusBadRequest,

	// 401 Unauthorized
	ERROR_UNAUTHORIZED: http.StatusUnauthorized,

	// 403 Forbidden

	// 404 Not Found
	ERROR_NOT_FOUND: http.StatusNotFound,

	// 409 Conflict
	ERROR_DUPLICATE_ENTITY:      http.StatusConflict,
	ERROR_UNIQUE_VIOLATION:      http.StatusConflict,
	ERROR_FOREIGN_KEY_VIOLATION: http.StatusConflict,

	// 429 Too Many Requests
	ERROR_TOO_MANY_REQUEST: http.StatusTooManyRequests,

	// 500 Internal Server Error
	ERROR_INTERNAL:                http.StatusInternalServerError,
	ERROR_UNKNOWN_ERROR:           http.StatusInternalServerError,
	ERROR_FAILED_TO_HASH_PASSWORD: http.StatusInternalServerError,
	ERROR_HASH_TOKEN:              http.StatusInternalServerError,
}

func (e *AppError) MapErrorCodeToHTTPCode() int {
	if code, ok := errorCodeMap[e.Code]; ok {
		return code
	}

	return http.StatusInternalServerError
}
