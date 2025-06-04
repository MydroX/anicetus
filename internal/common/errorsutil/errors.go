package errorsutil

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Severity string

const (
	databaseUnavailableErrorMsg = "Database is unavailable"

	SeverityWarn     Severity = "WARN"     // For potential harmful situations
	SeverityError    Severity = "ERROR"    // For standard errors
	SeverityCritical Severity = "CRITICAL" // For severe errors that might impair the system
	SeverityFatal    Severity = "FATAL"    // For errors that are likely to cause the system to stop functioning
)

type AppError struct {
	Code     int
	Message  string
	Err      error
	Severity Severity
	Source   string
	TraceID  string
	// Timestamp time.Time
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code int, message string, err error) error {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func NotFound(id string) *AppError {
	return &AppError{
		Code:    ErrorNotFound,
		Message: fmt.Sprintf("Entity with ID %s not found", id),
	}
}

func DuplicateEntity(id string) *AppError {
	return &AppError{
		Code:    ErrorDuplicateEntity,
		Message: fmt.Sprintf("Entity with ID %s already exists", id),
	}
}

func FailToBind() *AppError {
	return &AppError{
		Code:    ErrorFailToBind,
		Message: "Failed to bind request",
	}
}

func FailToValidate() *AppError {
	return &AppError{
		Code:    ErrorInvalidInput,
		Message: "Failed to validate request",
	}
}

func InvalidUUID() *AppError {
	return &AppError{
		Code:    ErrorInvalidUUID,
		Message: "Invalid UUID",
	}
}

func InvalidUsername() *AppError {
	return &AppError{
		Code:    ErrorInvalidUsername,
		Message: "Invalid username",
	}
}

func FailedToHashPassword() *AppError {
	return &AppError{
		Code:    ErrorFailedToHashPassword,
		Message: "Failed to hash password",
	}
}

func InvalidPassword() *AppError {
	return &AppError{
		Code:    ErrorInvalidPassword,
		Message: "Invalid password",
	}
}

func SQLErrorParser(err error) error {
	var pgErr *pgconn.PgError

	if err == pgx.ErrNoRows {
		return &AppError{
			Code:    ErrorNotFound,
			Message: "Entity not found",
			Err:     err,
		}
	}

	if strings.Contains(err.Error(), "failed to connect") {
		return &AppError{
			Code:     ErrorDatabaseUnavailable,
			Message:  databaseUnavailableErrorMsg,
			Err:      err,
			Severity: SeverityCritical,
		}
	}

	if ok := errors.As(err, &pgErr); !ok {
		return &AppError{
			Code:     ErrorUnknownErrorDB,
			Message:  "Internal error from database",
			Err:      err,
			Severity: SeverityCritical,
		}
	}

	switch pgErr.Code {
	case pgerrcode.AdminShutdown, pgerrcode.CrashShutdown, pgerrcode.CannotConnectNow:
		return &AppError{
			Code:    ErrorDatabaseUnavailable,
			Message: pgErr.Message,
			Err:     err,
		}
	case pgerrcode.UniqueViolation:
		return &AppError{
			Code:    ErrorUniqueViolation,
			Message: pgErr.Message,
			Err:     err,
		}
	case pgerrcode.ForeignKeyViolation:
		return &AppError{
			Code:    ErrorForeignKeyViolation,
			Message: pgErr.Message,
			Err:     err,
		}
	case pgerrcode.CheckViolation:
		return &AppError{
			Code:    ErrorConstraintViolation,
			Message: pgErr.Message,
			Err:     err,
		}
	case pgerrcode.NotNullViolation:
		return &AppError{
			Code:    ErrorNotNullViolation,
			Message: pgErr.Message,
			Err:     err,
		}
	default:
		return &AppError{
			Code:     ErrorInternal,
			Message:  "Internal error from database",
			Err:      err,
			Severity: SeverityCritical,
		}
	}
}
