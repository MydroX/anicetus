package errs

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type AppError struct {
	Code    int
	Message string
	Err     error
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

func SQLErrorParser(err error) error {
	var pgErr *pgconn.PgError

	if err == pgx.ErrNoRows {
		return &AppError{Code: ErrorNotFound, Message: "entity not found", Err: err}
	}

	if strings.Contains(err.Error(), "failed to connect") {
		return &AppError{Code: ErrorDatabaseUnavailable, Message: "database is unavailable", Err: err}
	}

	if ok := errors.As(err, &pgErr); !ok {
		return &AppError{Code: ErrorUnknownErrorDB, Message: "internal database error", Err: err}
	}

	switch pgErr.Code {
	case pgerrcode.AdminShutdown, pgerrcode.CrashShutdown, pgerrcode.CannotConnectNow:
		return &AppError{Code: ErrorDatabaseUnavailable, Message: "database is unavailable", Err: err}
	case pgerrcode.UniqueViolation:
		return &AppError{Code: ErrorUniqueViolation, Message: "entity already exists", Err: err}
	case pgerrcode.ForeignKeyViolation:
		return &AppError{Code: ErrorForeignKeyViolation, Message: "referenced entity not found", Err: err}
	case pgerrcode.CheckViolation:
		return &AppError{Code: ErrorConstraintViolation, Message: "constraint violation", Err: err}
	case pgerrcode.NotNullViolation:
		return &AppError{Code: ErrorNotNullViolation, Message: "missing required field", Err: err}
	default:
		return &AppError{Code: ErrorInternal, Message: "internal database error", Err: err}
	}
}
