package errs

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	inner := errors.New("inner error")
	err := New(ErrorNotFound, "not found", inner)

	var appErr *AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, ErrorNotFound, appErr.Code)
	assert.Equal(t, "not found", appErr.Message)
	assert.Equal(t, inner, appErr.Err)
}

func TestAppError_Error(t *testing.T) {
	appErr := &AppError{Code: 10001, Message: "bind failed"}
	assert.Equal(t, "10001: bind failed", appErr.Error())
}

func TestAppError_Unwrap(t *testing.T) {
	inner := errors.New("root cause")
	appErr := &AppError{Err: inner}

	assert.Equal(t, inner, appErr.Unwrap())
	assert.True(t, errors.Is(appErr, inner))
}

func TestAppError_Unwrap_Nil(t *testing.T) {
	appErr := &AppError{}
	assert.Nil(t, appErr.Unwrap())
}

func TestSQLErrorParser(t *testing.T) {
	tests := map[string]struct {
		inputErr     error
		expectedCode int
		expectedMsg  string
	}{
		"no rows": {
			inputErr:     pgx.ErrNoRows,
			expectedCode: ErrorNotFound,
			expectedMsg:  "entity not found",
		},
		"failed to connect": {
			inputErr:     fmt.Errorf("failed to connect to host"),
			expectedCode: ErrorDatabaseUnavailable,
			expectedMsg:  "database is unavailable",
		},
		"non pg error": {
			inputErr:     fmt.Errorf("some random db error"),
			expectedCode: ErrorUnknownErrorDB,
			expectedMsg:  "internal database error",
		},
		"admin shutdown": {
			inputErr:     &pgconn.PgError{Code: pgerrcode.AdminShutdown},
			expectedCode: ErrorDatabaseUnavailable,
			expectedMsg:  "database is unavailable",
		},
		"crash shutdown": {
			inputErr:     &pgconn.PgError{Code: pgerrcode.CrashShutdown},
			expectedCode: ErrorDatabaseUnavailable,
			expectedMsg:  "database is unavailable",
		},
		"cannot connect now": {
			inputErr:     &pgconn.PgError{Code: pgerrcode.CannotConnectNow},
			expectedCode: ErrorDatabaseUnavailable,
			expectedMsg:  "database is unavailable",
		},
		"unique violation": {
			inputErr:     &pgconn.PgError{Code: pgerrcode.UniqueViolation},
			expectedCode: ErrorUniqueViolation,
			expectedMsg:  "entity already exists",
		},
		"foreign key violation": {
			inputErr:     &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation},
			expectedCode: ErrorForeignKeyViolation,
			expectedMsg:  "referenced entity not found",
		},
		"check violation": {
			inputErr:     &pgconn.PgError{Code: pgerrcode.CheckViolation},
			expectedCode: ErrorConstraintViolation,
			expectedMsg:  "constraint violation",
		},
		"not null violation": {
			inputErr:     &pgconn.PgError{Code: pgerrcode.NotNullViolation},
			expectedCode: ErrorNotNullViolation,
			expectedMsg:  "missing required field",
		},
		"unknown pg error code": {
			inputErr:     &pgconn.PgError{Code: "99999"},
			expectedCode: ErrorInternal,
			expectedMsg:  "internal database error",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := SQLErrorParser(tc.inputErr)

			var appErr *AppError
			require.True(t, errors.As(result, &appErr))
			assert.Equal(t, tc.expectedCode, appErr.Code)
			assert.Equal(t, tc.expectedMsg, appErr.Message)
			assert.NotNil(t, appErr.Err)
		})
	}
}

func TestMapErrorCodeToHTTPCode(t *testing.T) {
	tests := map[string]struct {
		code         int
		expectedHTTP int
	}{
		"fail to bind":           {ErrorFailToBind, http.StatusBadRequest},
		"invalid input":          {ErrorInvalidInput, http.StatusBadRequest},
		"invalid uuid":           {ErrorInvalidUUID, http.StatusBadRequest},
		"invalid username":       {ErrorInvalidUsername, http.StatusBadRequest},
		"invalid password":       {ErrorInvalidPassword, http.StatusBadRequest},
		"constraint violation":   {ErrorConstraintViolation, http.StatusBadRequest},
		"not null violation":     {ErrorNotNullViolation, http.StatusBadRequest},
		"unauthorized":           {ErrorUnauthorized, http.StatusUnauthorized},
		"invalid credentials":    {ErrorInvalidCredentials, http.StatusUnauthorized},
		"forbidden":              {ErrorForbidden, http.StatusForbidden},
		"not found":              {ErrorNotFound, http.StatusNotFound},
		"duplicate entity":       {ErrorDuplicateEntity, http.StatusConflict},
		"unique violation":       {ErrorUniqueViolation, http.StatusConflict},
		"foreign key violation":  {ErrorForeignKeyViolation, http.StatusConflict},
		"too many requests":      {ErrorTooManyRequest, http.StatusTooManyRequests},
		"internal":               {ErrorInternal, http.StatusInternalServerError},
		"unknown error":          {ErrorUnknownError, http.StatusInternalServerError},
		"failed to hash":         {ErrorFailedToHashPassword, http.StatusInternalServerError},
		"hash token":             {ErrorHashToken, http.StatusInternalServerError},
		"unmapped code defaults": {999999, http.StatusInternalServerError},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			appErr := &AppError{Code: tc.code}
			assert.Equal(t, tc.expectedHTTP, appErr.MapErrorCodeToHTTPCode())
		})
	}
}
