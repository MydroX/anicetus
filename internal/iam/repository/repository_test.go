package repository

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"MydroX/anicetus/internal/common/context"
	"MydroX/anicetus/internal/common/errorsutil"
	"MydroX/anicetus/internal/iam/models"
	"MydroX/anicetus/pkg/logger"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveSession(t *testing.T) {
	// Test setup
	ctx := context.NewAppContextTest()
	log := logger.New("TEST")

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err, "Failed to create mock pool")
	defer poolMock.Close()

	repo := New(log, poolMock)
	queries := &Queries{}

	// Get the actual query and escape for regex
	saveSessionQuery := regexp.QuoteMeta(queries.SaveSession())

	// Current time for consistent testing
	now := time.Now()
	expiryTime := now.Add(24 * time.Hour)

	// Test session data
	testSession := &models.Session{
		UUID:           "session_123e4567-e89b-12d3-a456-426614174000",
		UserId:         "user_123e4567-e89b-12d3-a456-426614174000",
		RefreshToken:   "refresh_token_123456",
		LastUsedAt:     now,
		OS:             "Windows",
		Browser:        "Chrome",
		BrowserVersion: "96.0.4664.93",
		IPv4Address:    "192.168.1.1",
		CreatedAt:      now,
		ExpiresAt:      expiryTime,
	}

	// Test cases
	tests := map[string]struct {
		session       *models.Session
		mockSetup     func()
		expectedError bool
		checkError    func(t *testing.T, err error)
	}{
		"Success": {
			session: testSession,
			mockSetup: func() {
				poolMock.ExpectExec(saveSessionQuery).
					WithArgs(
						testSession.UUID,
						testSession.UserId,
						testSession.RefreshToken,
						testSession.LastUsedAt,
						testSession.OS,
						testSession.Browser,
						testSession.BrowserVersion,
						testSession.IPv4Address,
						testSession.CreatedAt,
						testSession.ExpiresAt,
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
			checkError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		"Unique Violation Error": {
			session: testSession,
			mockSetup: func() {
				pgErr := &pgconn.PgError{
					Code:           pgerrcode.UniqueViolation,
					Message:        "duplicate key value violates unique constraint",
					Detail:         "Key (uuid)=(session_123e4567-e89b-12d3-a456-426614174000) already exists.",
					TableName:      "sessions",
					ConstraintName: "sessions_pkey",
				}

				poolMock.ExpectExec(saveSessionQuery).
					WithArgs(
						testSession.UUID,
						testSession.UserId,
						testSession.RefreshToken,
						testSession.LastUsedAt,
						testSession.OS,
						testSession.Browser,
						testSession.BrowserVersion,
						testSession.IPv4Address,
						testSession.CreatedAt,
						testSession.ExpiresAt,
					).
					WillReturnError(pgErr)
			},
			expectedError: true,
			checkError: func(t *testing.T, err error) {
				assert.Error(t, err)

				var appErr *errorsutil.AppError
				if assert.True(t, errors.As(err, &appErr), "Expected an AppError") {
					assert.Equal(t, errorsutil.ERROR_UNIQUE_VIOLATION, appErr.Code)
					assert.Contains(t, appErr.Message, "session: unique violation")
				}
			},
		},
		"Generic Database Error": {
			session: testSession,
			mockSetup: func() {
				// Create a generic PG error that isn't a unique violation
				pgErr := &pgconn.PgError{
					Code:    pgerrcode.CheckViolation,
					Message: "check constraint violation",
				}

				poolMock.ExpectExec(saveSessionQuery).
					WithArgs(
						testSession.UUID,
						testSession.UserId,
						testSession.RefreshToken,
						testSession.LastUsedAt,
						testSession.OS,
						testSession.Browser,
						testSession.BrowserVersion,
						testSession.IPv4Address,
						testSession.CreatedAt,
						testSession.ExpiresAt,
					).
					WillReturnError(pgErr)
			},
			expectedError: true,
			checkError: func(t *testing.T, err error) {
				assert.Error(t, err)

				var appErr *errorsutil.AppError
				if assert.True(t, errors.As(err, &appErr), "Expected an AppError") {
					assert.Equal(t, errorsutil.ERROR_INTERNAL, appErr.Code)
					assert.Contains(t, appErr.Message, "session: error during save")
				}
			},
		},
	}

	// Run all test cases
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup mock expectations
			tc.mockSetup()

			// Call the function being tested
			err := repo.SaveSession(ctx, tc.session)

			// Check results with the custom check function
			tc.checkError(t, err)
		})
	}

	// Ensure all expectations were met
	assert.NoError(t, poolMock.ExpectationsWereMet())
}
