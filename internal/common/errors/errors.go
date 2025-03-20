package errors

import "fmt"

type Err Error

type Error struct {
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(code, message string, err error) *Err {
	return &Err{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func NotFound(id string) *Error {
	return &Error{
		Code:    ERROR_NOT_FOUND,
		Message: fmt.Sprintf("Entity with ID %s not found", id),
	}
}

func DuplicateEntity(id string) *Error {
	return &Error{
		Code:    ERROR_DUPLICATE_ENTITY,
		Message: fmt.Sprintf("Entity with ID %s already exists", id),
	}
}

func FailToBind() *Error {
	return &Error{
		Code:    ERROR_FAIL_TO_BIND,
		Message: "Failed to bind request",
	}
}

func FailToValidate() *Error {
	return &Error{
		Code:    ERROR_INVALID_INPUT,
		Message: "Failed to validate request",
	}
}

func InvalidUUID() *Error {
	return &Error{
		Code:    ERROR_INVALID_UUID,
		Message: "Invalid UUID",
	}
}

func InvalidUsername() *Error {
	return &Error{
		Code:    ERROR_INVALID_USERNAME,
		Message: "Invalid username",
	}
}

func FailedToHashPassword() *Error {
	return &Error{
		Code:    ERROR_FAILED_TO_HASH_PASSWORD,
		Message: "Failed to hash password",
	}
}

func InvalidPassword() *Error {
	return &Error{
		Code:    ERROR_INVALID_PASSWORD,
		Message: "Invalid password",
	}
}
