package errorsutil

import "net/http"

// Error codes with numeric values
// Format: CCXXXX where:
// - CC: Category code (2 digits)
// - XXXX: Error code (4 digits)
const (

	// Common errors (10xxx)
	ERROR_FAIL_TO_BIND     = 10001
	ERROR_INVALID_INPUT    = 10002
	ERROR_INVALID_UUID     = 10003
	ERROR_NOT_FOUND        = 10004
	ERROR_DUPLICATE_ENTITY = 10005
	ERROR_UNAUTHORIZED     = 10006
	ERROR_TOO_MANY_REQUEST = 10007

	ERROR_INTERNAL      = 10008
	ERROR_UNKNOWN_ERROR = 10999

	// User errors (11xxx)
	ERROR_INVALID_USERNAME        = 11001
	ERROR_FAILED_TO_HASH_PASSWORD = 11002
	ERROR_INVALID_PASSWORD        = 11003

	// IAM errors (12xxx)
	ERROR_HASH_TOKEN = 12001

	// Database errors (99xxx)
	ERROR_UNIQUE_VIOLATION      = 99001
	ERROR_FOREIGN_KEY_VIOLATION = 99002
	ERROR_CHECK_VIOLATION       = 99003
	ERROR_NOT_NULL_VIOLATION    = 99004
	ERROR_DATABASE_UNAVAILABLE  = 99005
	ERROR_UNKNOWN_ERROR_DB      = 99999
)

const (
	MessageFailToBind   = "Failed to bind request. Please check your request and try again"
	MessageInvalidInput = "Invalid input. Please check your request and try again"
	MessageInvalidUUID  = "Invalid UUID"
)

var errorCodeMap = map[int]int{
	// 400 Bad Request
	ERROR_FAIL_TO_BIND:       http.StatusBadRequest,
	ERROR_INVALID_INPUT:      http.StatusBadRequest,
	ERROR_INVALID_UUID:       http.StatusBadRequest,
	ERROR_INVALID_USERNAME:   http.StatusBadRequest,
	ERROR_INVALID_PASSWORD:   http.StatusBadRequest,
	ERROR_CHECK_VIOLATION:    http.StatusBadRequest,
	ERROR_NOT_NULL_VIOLATION: http.StatusBadRequest,

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
