package errors

import "errors"

type errorCodesKey string

const (
	// Common errors
	CODE_INVALID_REQUEST  = "00001"
	CODE_INVALID_UUID     = "00002"
	CODE_ENTITY_NOT_FOUND = "00003"
	CODE_DUPLICATE_ENTITY = "00004"
	CODE_UNKNOWN_ERROR    = "00999"

	// Users errors
	CODE_INVALID_USERNAME        = "01001"
	CODE_FAILED_TO_HASH_PASSWORD = "01002"
)

var (
	CtxErrorCodeKey = errorCodesKey("error_code")

	ErrNotFound = errors.New("entity not found")
)
