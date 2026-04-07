package repository

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"context"
	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/internal/identity/models"
	"MydroX/anicetus/pkg/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	// Test setup
	ctx := context.Background()
	log, err := logger.New("TEST")
	if err != nil {
		panic(err)
	}
	queries := &IdentityQueries{}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err, "Failed to create mock pool")
	defer poolMock.Close()

	repo := New(log, poolMock)
	query := regexp.QuoteMeta(queries.CreateUser())

	// Test user data
	testUser := &models.User{
		UUID:     "user_123e4567-e89b-12d3-a456-426614174000",
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword123",
	}

	// Test cases
	tests := []struct {
		name          string
		mockSetup     func()
		expectedError string
		errorCode     int
	}{
		{
			name: "Success",
			mockSetup: func() {
				poolMock.ExpectExec(query).
					WithArgs(testUser.UUID, testUser.Username, testUser.Email, testUser.Password).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: "",
		},
		{
			name: "Unique violation error",
			mockSetup: func() {
				pgErr := &pgconn.PgError{
					Code:           "23505",
					Message:        "duplicate key value violates unique constraint",
					Detail:         "Key (email)=(test@example.com) already exists.",
					TableName:      "users",
					ConstraintName: "users_email_key",
				}
				poolMock.ExpectExec(query).
					WithArgs(testUser.UUID, testUser.Username, testUser.Email, testUser.Password).
					WillReturnError(pgErr)
			},
			expectedError: "entity already exists",
			errorCode:     errs.ErrorUniqueViolation,
		},
		{
			name: "Not null violation error",
			mockSetup: func() {
				pgErr := &pgconn.PgError{
					Code:           "23502",
					Message:        "null value in column \"email\" violates not-null constraint",
					Detail:         "Failing row contains (user_id, testuser, null, hashedpassword123, USER).",
					TableName:      "users",
					ConstraintName: "users_email_not_null",
				}
				poolMock.ExpectExec(query).
					WithArgs(testUser.UUID, testUser.Username, testUser.Email, testUser.Password).
					WillReturnError(pgErr)
			},
			expectedError: "missing required field",
			errorCode:     errs.ErrorNotNullViolation,
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup expectations for this test case
			tc.mockSetup()

			// Execute the function being tested
			err := repo.CreateUser(ctx, testUser)

			// Verify expectations
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)

				// Check for specific error type if applicable
				if tc.errorCode != 0 {
					var appErr *errs.AppError
					assert.True(t, errors.As(err, &appErr), "Expected an AppError")
					assert.Equal(t, tc.errorCode, appErr.Code)
				}
			}
		})
	}

	// Ensure all expectations were met
	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestGetUserByUUID(t *testing.T) {
	// Test setup
	ctx := context.Background()
	log, err := logger.New("TEST")
	if err != nil {
		panic(err)
	}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err, "Failed to create mock pool")
	defer poolMock.Close()

	repo := New(log, poolMock)
	queries := &IdentityQueries{}

	// Get the actual query from your function and escape for regex
	actualQuery := queries.GetUserByUUID()
	query := regexp.QuoteMeta(actualQuery)

	// Fixed test data
	testUUID := "user_123e4567-e89b-12d3-a456-426614174000"
	now := time.Now()

	// Complete user object for successful return
	expectedUser := &models.User{
		UUID:      testUUID,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword123",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Define test cases using map pattern
	tests := map[string]struct {
		uuid            string
		mockSetup       func()
		expectedUser    *models.User
		expectedError   bool
		expectedErrCode int
		checkFunc       func(t *testing.T, user *models.User, err error)
	}{
		"Success": {
			uuid: testUUID,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"uuid", "username", "email", "password", "created_at", "updated_at"}).
					AddRow(expectedUser.UUID, expectedUser.Username, expectedUser.Email, expectedUser.Password,
						expectedUser.CreatedAt, expectedUser.UpdatedAt)

				poolMock.ExpectQuery(query).
					WithArgs(testUUID).
					WillReturnRows(rows)
			},
			expectedUser:    expectedUser,
			expectedError:   false,
			expectedErrCode: 0,
			checkFunc: func(t *testing.T, user *models.User, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, expectedUser.UUID, user.UUID)
				assert.Equal(t, expectedUser.Username, user.Username)
				assert.Equal(t, expectedUser.Email, user.Email)
				assert.Equal(t, expectedUser.Password, user.Password)
				assert.Equal(t, expectedUser.CreatedAt.Truncate(time.Second),
					user.CreatedAt.Truncate(time.Second))
				assert.Equal(t, expectedUser.UpdatedAt.Truncate(time.Second),
					user.UpdatedAt.Truncate(time.Second))
			},
		},
		"User Not Found": {
			uuid: testUUID,
			mockSetup: func() {
				poolMock.ExpectQuery(query).
					WithArgs(testUUID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedUser:    nil,
			expectedError:   true,
			expectedErrCode: errs.ErrorNotFound,
			checkFunc: func(t *testing.T, user *models.User, err error) {
				assert.Error(t, err)
				assert.Nil(t, user)

				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr), "Expected an AppError") {
					assert.Equal(t, errs.ErrorNotFound, appErr.Code)
				}
			},
		},
	}

	// Run all the test cases
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup expectations for this test case
			tc.mockSetup()

			// Call the function being tested
			user, err := repo.GetUserByUUID(ctx, tc.uuid)

			// Run specific checks for this test case
			tc.checkFunc(t, user, err)
		})
	}

	// Ensure all expectations were met
	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestUpdateUser(t *testing.T) {
	// Test setup
	ctx := context.Background()
	log, err := logger.New("TEST")
	if err != nil {
		panic(err)
	}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err, "Failed to create mock pool")
	defer poolMock.Close()

	repo := New(log, poolMock)
	queries := &IdentityQueries{}

	// Get the actual query and escape for regex
	updateQuery := regexp.QuoteMeta(queries.UpdateUser())

	// Test data
	testUUID := "user_123e4567-e89b-12d3-a456-426614174000"
	now := time.Now()

	inputUser := &models.User{
		UUID:     testUUID,
		Username: "updated_username",
		Email:    "updated@example.com",
	}

	// User with updated fields and additional fields from DB
	updatedUser := &models.User{
		UUID:      testUUID,
		Username:  "updated_username",
		Email:     "updated@example.com",
		Password:  "hashedpassword123",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test cases
	tests := map[string]struct {
		user            *models.User
		mockSetup       func()
		expectedUser    *models.User
		expectedError   bool
		expectedErrMsg  string
		expectedErrCode int
	}{
		"Success": {
			user: inputUser,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"uuid", "username", "email", "password", "created_at", "updated_at"}).
					AddRow(updatedUser.UUID, updatedUser.Username, updatedUser.Email, updatedUser.Password,
						updatedUser.CreatedAt, updatedUser.UpdatedAt)

				poolMock.ExpectQuery(updateQuery).
					WithArgs(inputUser.Username, inputUser.Email, inputUser.UUID).
					WillReturnRows(rows)
			},
			expectedUser:    updatedUser,
			expectedError:   false,
			expectedErrCode: 0,
		},
		"User Not Found": {
			user: inputUser,
			mockSetup: func() {
				poolMock.ExpectQuery(updateQuery).
					WithArgs(inputUser.Username, inputUser.Email, inputUser.UUID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedUser:    nil,
			expectedError:   true,
			expectedErrMsg:  "no rows",
			expectedErrCode: errs.ErrorNotFound,
		},
		"Unique Violation": {
			user: inputUser,
			mockSetup: func() {
				pgErr := &pgconn.PgError{
					Code:           "23505",
					Message:        "duplicate key value violates unique constraint",
					Detail:         "Key (email)=(updated@example.com) already exists.",
					TableName:      "users",
					ConstraintName: "users_email_key",
				}
				poolMock.ExpectQuery(updateQuery).
					WithArgs(inputUser.Username, inputUser.Email, inputUser.UUID).
					WillReturnError(pgErr)
			},
			expectedUser:    nil,
			expectedError:   true,
			expectedErrMsg:  "entity already exists",
			expectedErrCode: errs.ErrorUniqueViolation,
		},
	}

	// Run all test cases
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup mock expectations
			tc.mockSetup()

			// Call the function being tested
			user, err := repo.UpdateUser(ctx, tc.user)

			// Check results
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, user)

				// Check error type if specific code expected
				if tc.expectedErrCode != 0 {
					var appErr *errs.AppError
					if assert.True(t, errors.As(err, &appErr), "Expected an AppError") {
						assert.Equal(t, tc.expectedErrCode, appErr.Code)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tc.expectedUser.UUID, user.UUID)
				assert.Equal(t, tc.expectedUser.Username, user.Username)
				assert.Equal(t, tc.expectedUser.Email, user.Email)
				assert.Equal(t, tc.expectedUser.Password, user.Password)
				assert.Equal(t, tc.expectedUser.CreatedAt.Truncate(time.Second),
					user.CreatedAt.Truncate(time.Second))
				assert.Equal(t, tc.expectedUser.UpdatedAt.Truncate(time.Second),
					user.UpdatedAt.Truncate(time.Second))
			}
		})
	}

	// Ensure all expectations were met
	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestUpdatePassword(t *testing.T) {
	// Test setup
	ctx := context.Background()
	log, err := logger.New("TEST")
	if err != nil {
		panic(err)
	}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err, "Failed to create mock pool")
	defer poolMock.Close()

	repo := New(log, poolMock)
	queries := &IdentityQueries{}

	// Get the actual query and escape for regex
	updateQuery := regexp.QuoteMeta(queries.UpdatePassword())

	// Test data
	testUUID := "user_123e4567-e89b-12d3-a456-426614174000"
	newPassword := "newhashedpassword456"

	// Test cases
	tests := map[string]struct {
		uuid            string
		password        string
		mockSetup       func()
		expectedError   bool
		expectedErrMsg  string
		expectedErrCode int
	}{
		"Success": {
			uuid:     testUUID,
			password: newPassword,
			mockSetup: func() {
				poolMock.ExpectExec(updateQuery).
					WithArgs(newPassword, testUUID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError:   false,
			expectedErrCode: 0,
		},
		"User Not Found": {
			uuid:     testUUID,
			password: newPassword,
			mockSetup: func() {
				poolMock.ExpectExec(updateQuery).
					WithArgs(newPassword, testUUID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError:   true,
			expectedErrMsg:  "user not found",
			expectedErrCode: errs.ErrorNotFound,
		},
		"Database Error": {
			uuid:     testUUID,
			password: newPassword,
			mockSetup: func() {
				pgErr := &pgconn.PgError{
					Code:    "57P03",
					Message: "database connection error",
				}
				poolMock.ExpectExec(updateQuery).
					WithArgs(newPassword, testUUID).
					WillReturnError(pgErr)
			},
			expectedError:   true,
			expectedErrMsg:  "database is unavailable",
			expectedErrCode: errs.ErrorDatabaseUnavailable,
		},
	}

	// Run all test cases
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup mock expectations
			tc.mockSetup()

			// Call the function being tested
			err := repo.UpdatePassword(ctx, tc.uuid, tc.password)

			// Check results
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)

				// Check error type if specific code expected
				if tc.expectedErrCode != 0 {
					var appErr *errs.AppError
					if assert.True(t, errors.As(err, &appErr), "Expected an AppError") {
						assert.Equal(t, tc.expectedErrCode, appErr.Code)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Ensure all expectations were met
	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestUpdateEmail(t *testing.T) {
	// Test setup
	ctx := context.Background()
	log, err := logger.New("TEST")
	if err != nil {
		panic(err)
	}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err, "Failed to create mock pool")
	defer poolMock.Close()

	repo := New(log, poolMock)
	queries := &IdentityQueries{}

	// Get the actual query and escape for regex
	updateQuery := regexp.QuoteMeta(queries.UpdateEmail())

	// Test data
	testUUID := "user_123e4567-e89b-12d3-a456-426614174000"
	newEmail := "newemail@example.com"

	// Test cases
	tests := map[string]struct {
		uuid            string
		email           string
		mockSetup       func()
		expectedError   bool
		expectedErrMsg  string
		expectedErrCode int
	}{
		"Success": {
			uuid:  testUUID,
			email: newEmail,
			mockSetup: func() {
				poolMock.ExpectExec(updateQuery).
					WithArgs(newEmail, testUUID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError:   false,
			expectedErrCode: 0,
		},
		"User Not Found": {
			uuid:  testUUID,
			email: newEmail,
			mockSetup: func() {
				poolMock.ExpectExec(updateQuery).
					WithArgs(newEmail, testUUID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError:   true,
			expectedErrMsg:  "user not found",
			expectedErrCode: errs.ErrorNotFound,
		},
		"Unique Violation": {
			uuid:  testUUID,
			email: newEmail,
			mockSetup: func() {
				pgErr := &pgconn.PgError{
					Code:           "23505",
					Message:        "duplicate key value violates unique constraint",
					Detail:         "Key (email)=(newemail@example.com) already exists.",
					TableName:      "users",
					ConstraintName: "users_email_key",
				}
				poolMock.ExpectExec(updateQuery).
					WithArgs(newEmail, testUUID).
					WillReturnError(pgErr)
			},
			expectedError:   true,
			expectedErrMsg:  "entity already exists",
			expectedErrCode: errs.ErrorUniqueViolation,
		},
	}

	// Run all test cases
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup mock expectations
			tc.mockSetup()

			// Call the function being tested
			err := repo.UpdateEmail(ctx, tc.uuid, tc.email)

			// Check results
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)

				// Check error type if specific code expected
				if tc.expectedErrCode != 0 {
					var appErr *errs.AppError
					if assert.True(t, errors.As(err, &appErr), "Expected an AppError") {
						assert.Equal(t, tc.expectedErrCode, appErr.Code)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Ensure all expectations were met
	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestDeleteUser(t *testing.T) {
	// Test setup
	ctx := context.Background()
	log, err := logger.New("TEST")
	if err != nil {
		panic(err)
	}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err, "Failed to create mock pool")
	defer poolMock.Close()

	repo := New(log, poolMock)
	queries := &IdentityQueries{}

	// Get the actual query and escape for regex
	deleteQuery := regexp.QuoteMeta(queries.DeleteUser())

	// Test data
	testUUID := "user_123e4567-e89b-12d3-a456-426614174000"

	// Test cases
	tests := map[string]struct {
		uuid          string
		mockSetup     func()
		expectedError bool
		checkError    func(t *testing.T, err error)
	}{
		"Success": {
			uuid: testUUID,
			mockSetup: func() {
				poolMock.ExpectExec(deleteQuery).
					WithArgs(testUUID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectedError: false,
			checkError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		"Database Error": {
			uuid: testUUID,
			mockSetup: func() {
				pgErr := &pgconn.PgError{
					Code:    "57P03",
					Message: "database connection error",
				}
				poolMock.ExpectExec(deleteQuery).
					WithArgs(testUUID).
					WillReturnError(pgErr)
			},
			expectedError: true,
			checkError: func(t *testing.T, err error) {
				var appErr *errs.AppError
				errors.As(err, &appErr)

				assert.Error(t, err)
				assert.Contains(t, err.Error(), "database is unavailable")
				assert.Equal(t, errs.ErrorDatabaseUnavailable, appErr.Code)
			},
		},
		"User Not Found": {
			uuid: testUUID,
			mockSetup: func() {
				poolMock.ExpectExec(deleteQuery).
					WithArgs(testUUID).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			expectedError: true,
			checkError: func(t *testing.T, err error) {
				assert.Error(t, err)
				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr), "Expected an AppError") {
					assert.Equal(t, errs.ErrorNotFound, appErr.Code)
					assert.Contains(t, appErr.Message, "user not found")
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
			err := repo.DeleteUser(ctx, tc.uuid)

			// Use the custom check function
			tc.checkError(t, err)
		})
	}

	// Ensure all expectations were met
	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestGetUserByEmail(t *testing.T) {
	// Test setup
	ctx := context.Background()
	log, err := logger.New("TEST")
	if err != nil {
		panic(err)
	}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err, "Failed to create mock pool")
	defer poolMock.Close()

	repo := New(log, poolMock)
	queries := &IdentityQueries{}

	// Get the actual query and escape for regex
	emailQuery := regexp.QuoteMeta(queries.GetUserByEmail())

	// Test data
	testEmail := "test@example.com"
	now := time.Now()

	// User data for successful return
	expectedUser := &models.User{
		UUID:      "user_123e4567-e89b-12d3-a456-426614174000",
		Username:  "testuser",
		Email:     testEmail,
		Password:  "hashedpassword123",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test cases
	tests := map[string]struct {
		email         string
		mockSetup     func()
		expectedUser  *models.User
		expectedError bool
		checkResult   func(t *testing.T, user *models.User, err error)
	}{
		"Success": {
			email: testEmail,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"uuid", "username", "email", "password", "created_at", "updated_at"}).
					AddRow(expectedUser.UUID, expectedUser.Username, expectedUser.Email, expectedUser.Password,
						expectedUser.CreatedAt, expectedUser.UpdatedAt)

				poolMock.ExpectQuery(emailQuery).
					WithArgs(testEmail).
					WillReturnRows(rows)
			},
			expectedUser:  expectedUser,
			expectedError: false,
			checkResult: func(t *testing.T, user *models.User, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, expectedUser.UUID, user.UUID)
				assert.Equal(t, expectedUser.Username, user.Username)
				assert.Equal(t, expectedUser.Email, user.Email)
				assert.Equal(t, expectedUser.Password, user.Password)
				assert.Equal(t, expectedUser.CreatedAt.Truncate(time.Second),
					user.CreatedAt.Truncate(time.Second))
				assert.Equal(t, expectedUser.UpdatedAt.Truncate(time.Second),
					user.UpdatedAt.Truncate(time.Second))
			},
		},
		"User Not Found": {
			email: testEmail,
			mockSetup: func() {
				poolMock.ExpectQuery(emailQuery).
					WithArgs(testEmail).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: true,
			checkResult: func(t *testing.T, user *models.User, err error) {
				assert.Error(t, err)
				assert.Nil(t, user)

				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr), "Expected an AppError") {
					assert.Equal(t, errs.ErrorNotFound, appErr.Code)
					assert.Contains(t, appErr.Message, "not found")
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
			user, err := repo.GetUserByEmail(ctx, tc.email)

			// Check results with the custom check function
			tc.checkResult(t, user, err)
		})
	}

	// Ensure all expectations were met
	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestGetUserByUsername(t *testing.T) {
	// Test setup
	ctx := context.Background()
	log, err := logger.New("TEST")
	if err != nil {
		panic(err)
	}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err, "Failed to create mock pool")
	defer poolMock.Close()

	repo := New(log, poolMock)
	queries := &IdentityQueries{}

	// Get the actual query and escape for regex
	usernameQuery := regexp.QuoteMeta(queries.GetUserByUsername())

	// Test data
	testUsername := "testuser"
	now := time.Now()

	// User data for successful return
	expectedUser := &models.User{
		UUID:      "user_123e4567-e89b-12d3-a456-426614174000",
		Username:  testUsername,
		Email:     "test@example.com",
		Password:  "hashedpassword123",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test cases
	tests := map[string]struct {
		username      string
		mockSetup     func()
		expectedUser  *models.User
		expectedError bool
		checkResult   func(t *testing.T, user *models.User, err error)
	}{
		"Success": {
			username: testUsername,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"uuid", "username", "email", "password", "created_at", "updated_at"}).
					AddRow(expectedUser.UUID, expectedUser.Username, expectedUser.Email, expectedUser.Password,
						expectedUser.CreatedAt, expectedUser.UpdatedAt)

				poolMock.ExpectQuery(usernameQuery).
					WithArgs(testUsername).
					WillReturnRows(rows)
			},
			expectedUser:  expectedUser,
			expectedError: false,
			checkResult: func(t *testing.T, user *models.User, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, expectedUser.UUID, user.UUID)
				assert.Equal(t, expectedUser.Username, user.Username)
				assert.Equal(t, expectedUser.Email, user.Email)
				assert.Equal(t, expectedUser.Password, user.Password)
				assert.Equal(t, expectedUser.CreatedAt.Truncate(time.Second),
					user.CreatedAt.Truncate(time.Second))
				assert.Equal(t, expectedUser.UpdatedAt.Truncate(time.Second),
					user.UpdatedAt.Truncate(time.Second))
			},
		},
		"User Not Found": {
			username: testUsername,
			mockSetup: func() {
				poolMock.ExpectQuery(usernameQuery).
					WithArgs(testUsername).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: true,
			checkResult: func(t *testing.T, user *models.User, err error) {
				assert.Error(t, err)
				assert.Nil(t, user)

				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr), "Expected an AppError") {
					assert.Equal(t, errs.ErrorNotFound, appErr.Code)
					assert.Contains(t, appErr.Message, "not found")
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
			user, err := repo.GetUserByUsername(ctx, tc.username)

			// Check results with the custom check function
			tc.checkResult(t, user, err)
		})
	}

	// Ensure all expectations were met
	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestGetAllUsers(t *testing.T) {
	// Test setup
	ctx := context.Background()
	log, err := logger.New("TEST")
	if err != nil {
		panic(err)
	}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err, "Failed to create mock pool")
	defer poolMock.Close()

	repo := New(log, poolMock)
	queries := &IdentityQueries{}

	// Get the actual query and escape for regex
	getAllQuery := regexp.QuoteMeta(queries.GetAllUsers())

	// Fixed time for consistent testing
	now := time.Now()

	// Sample users for test data
	user1 := &models.User{
		UUID:      "user_123e4567-e89b-12d3-a456-426614174000",
		Username:  "user1",
		Email:     "user1@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	user2 := &models.User{
		UUID:      "user_223e4567-e89b-12d3-a456-426614174000",
		Username:  "user2",
		Email:     "user2@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test cases
	tests := map[string]struct {
		mockSetup     func()
		expectedUsers []*models.User
		expectedError bool
		checkResult   func(t *testing.T, users []*models.User, err error)
	}{
		"Success Multiple Users": {
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"uuid", "username", "email", "created_at", "updated_at"}).
					AddRow(user1.UUID, user1.Username, user1.Email, user1.CreatedAt, user1.UpdatedAt).
					AddRow(user2.UUID, user2.Username, user2.Email, user2.CreatedAt, user2.UpdatedAt)

				poolMock.ExpectQuery(getAllQuery).
					WillReturnRows(rows)
			},
			expectedUsers: []*models.User{user1, user2},
			expectedError: false,
			checkResult: func(t *testing.T, users []*models.User, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, users)
				assert.Len(t, users, 2)

				// Check first user
				assert.Equal(t, user1.UUID, users[0].UUID)
				assert.Equal(t, user1.Username, users[0].Username)
				assert.Equal(t, user1.Email, users[0].Email)

				// Check second user
				assert.Equal(t, user2.UUID, users[1].UUID)
				assert.Equal(t, user2.Username, users[1].Username)
				assert.Equal(t, user2.Email, users[1].Email)
			},
		},
		"Success Empty Result": {
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"uuid", "username", "email", "created_at", "updated_at"})

				poolMock.ExpectQuery(getAllQuery).
					WillReturnRows(rows)
			},
			expectedUsers: make([]*models.User, 0),
			expectedError: false,
			checkResult: func(t *testing.T, users []*models.User, err error) {
				assert.NoError(t, err)
				assert.Len(t, users, 0)
				assert.Empty(t, users)
			},
		},
		"Query Error": {
			mockSetup: func() {
				poolMock.ExpectQuery(getAllQuery).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedUsers: nil,
			expectedError: true,
			checkResult: func(t *testing.T, users []*models.User, err error) {
				assert.Error(t, err)
				assert.Nil(t, users)
				var appErr *errs.AppError
				if assert.ErrorAs(t, err, &appErr) {
					assert.Equal(t, errs.ErrorNotFound, appErr.Code)
				}
			},
		},
		"Error Scanning User": {
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"uuid", "username", "email", "created_at", "updated_at"}).
					AddRow(user1.UUID, user1.Username, user1.Email, user1.CreatedAt, user1.UpdatedAt).
					AddRow(user2.UUID, user2.Username, user2.Email, user2.CreatedAt, user2.UpdatedAt).
					RowError(1, errors.New("error scanning user"))

				poolMock.ExpectQuery(getAllQuery).
					WillReturnRows(rows)
			},
			expectedUsers: nil,
			expectedError: true,
			checkResult: func(t *testing.T, users []*models.User, err error) {
				assert.Error(t, err)
				assert.Nil(t, users)
			},
		},
	}

	// Run all test cases
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup mock expectations
			tc.mockSetup()

			// Call the function being tested
			users, err := repo.GetAllUsers(ctx)

			// Check results with the custom check function
			tc.checkResult(t, users, err)
		})
	}

	// Ensure all expectations were met
	assert.NoError(t, poolMock.ExpectationsWereMet())
}
