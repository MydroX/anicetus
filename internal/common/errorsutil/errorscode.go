package errorsutil

import "net/http"

// Error codes with numeric values
// Format: CCXXXX where:
// - CC: Category code (2 digits)
// - XXXX: Error code (4 digits)
const (

	// Common errors (10xxx)
	ErrorFailToBind      = 10001 // Failed to bind request
	ErrorInvalidInput    = 10002 // Invalid input
	ErrorInvalidUUID     = 10003 // Invalid UUID
	ErrorNotFound        = 10004 // Not found
	ErrorDuplicateEntity = 10005 // Duplicate entity
	ErrorUnauthorized    = 10006 // Unauthorized
	ErrorTooManyRequest  = 10007 // Too many requests

	ErrorInternal     = 10008 // Internal server error
	ErrorUnknownError = 10999 // Unknown error

	// User errors (11xxx)
	ErrorInvalidUsername      = 11001 // Username does not meet the requirements
	ErrorFailedToHashPassword = 11002 // Failed to hash password
	ErrorInvalidPassword      = 11003 // Password does not meet the requirements

	// IAM errors (12xxx)
	ErrorInvalidCredentials = 12001 // Invalid credentials
	ErrorHashToken          = 12002 // Failed to hash token
	ErrorCreateToken        = 12003 // Failed to create token

	// Database errors (99xxx)
	ErrorUniqueViolation     = 99001 // Unique constraint violation
	ErrorForeignKeyViolation = 99002 // Foreign key constraint violation
	ErrorConstraintViolation = 99003 // Check constraint violation
	ErrorNotNullViolation    = 99004 // Not null constraint violation
	ErrorDatabaseUnavailable = 99005 // Database unavailable
	ErrorUnknownErrorDB      = 99999 // Unknown database error
)

const (
	MessageFailToBind   = "Failed to bind request. Please check your request and try again"
	MessageInvalidInput = "Invalid input. Please check your request and try again"
	MessageInvalidUUID         = "Invalid UUID"
	MessageInvalidCredentials  = "Invalid credentials"
)

var errorCodeMap = map[int]int{
	// 400 Bad Request
	ErrorFailToBind:          http.StatusBadRequest,
	ErrorInvalidInput:        http.StatusBadRequest,
	ErrorInvalidUUID:         http.StatusBadRequest,
	ErrorInvalidUsername:     http.StatusBadRequest,
	ErrorInvalidPassword:     http.StatusBadRequest,
	ErrorConstraintViolation: http.StatusBadRequest,
	ErrorNotNullViolation:    http.StatusBadRequest,

	// 401 Unauthorized
	ErrorUnauthorized:       http.StatusUnauthorized,
	ErrorInvalidCredentials: http.StatusUnauthorized,

	// 403 Forbidden

	// 404 Not Found
	ErrorNotFound: http.StatusNotFound,

	// 409 Conflict
	ErrorDuplicateEntity:     http.StatusConflict,
	ErrorUniqueViolation:     http.StatusConflict,
	ErrorForeignKeyViolation: http.StatusConflict,

	// 429 Too Many Requests
	ErrorTooManyRequest: http.StatusTooManyRequests,

	// 500 Internal Server Error
	ErrorInternal:             http.StatusInternalServerError,
	ErrorUnknownError:         http.StatusInternalServerError,
	ErrorFailedToHashPassword: http.StatusInternalServerError,
	ErrorHashToken:            http.StatusInternalServerError,
}

func (e *AppError) MapErrorCodeToHTTPCode() int {
	if code, ok := errorCodeMap[e.Code]; ok {
		return code
	}

	return http.StatusInternalServerError
}
