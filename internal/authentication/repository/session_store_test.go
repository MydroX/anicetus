package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"MydroX/anicetus/internal/authentication/models"
	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSession() *models.Session {
	now := time.Now().Truncate(time.Second)
	return &models.Session{
		UUID:           "sess-uuid-123",
		UserID:         "user-uuid-456",
		RefreshToken:   "$argon2id$v=19$m=65536,t=3,p=2$hash",
		LastUsedAt:     now,
		OS:             "macOS",
		OSVersion:      "14.0",
		Browser:        "Chrome",
		BrowserVersion: "120.0",
		IPv4Address:    "192.168.1.1",
		CreatedAt:      now,
		ExpiresAt:      now.Add(24 * time.Hour),
	}
}

func sessionColumns() []string {
	return []string{
		"uuid", "user_uuid", "refresh_token_hash", "last_used_at",
		"os", "os_version", "browser", "browser_version",
		"ipv4_address", "created_at", "expires_at",
	}
}

func TestSaveSession(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("TEST")
	queries := &SessionQueries{}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	repo := New(log, poolMock)
	query := regexp.QuoteMeta(queries.SaveSession())
	session := testSession()

	t.Run("success", func(t *testing.T) {
		poolMock.ExpectExec(query).
			WithArgs(
				session.UUID, session.UserID, session.RefreshToken,
				session.LastUsedAt, session.OS, session.OSVersion,
				session.Browser, session.BrowserVersion, session.IPv4Address,
				session.CreatedAt, session.ExpiresAt,
			).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.SaveSession(ctx, session)
		assert.NoError(t, err)
		assert.NoError(t, poolMock.ExpectationsWereMet())
	})

	t.Run("unique violation", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23505", Message: "duplicate key"}
		poolMock.ExpectExec(query).
			WithArgs(
				session.UUID, session.UserID, session.RefreshToken,
				session.LastUsedAt, session.OS, session.OSVersion,
				session.Browser, session.BrowserVersion, session.IPv4Address,
				session.CreatedAt, session.ExpiresAt,
			).
			WillReturnError(pgErr)

		err := repo.SaveSession(ctx, session)
		require.Error(t, err)

		var appErr *errs.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, errs.ErrorUniqueViolation, appErr.Code)
		assert.NoError(t, poolMock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		poolMock.ExpectExec(query).
			WithArgs(
				session.UUID, session.UserID, session.RefreshToken,
				session.LastUsedAt, session.OS, session.OSVersion,
				session.Browser, session.BrowserVersion, session.IPv4Address,
				session.CreatedAt, session.ExpiresAt,
			).
			WillReturnError(errors.New("failed to connect to host"))

		err := repo.SaveSession(ctx, session)
		require.Error(t, err)

		var appErr *errs.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, errs.ErrorDatabaseUnavailable, appErr.Code)
		assert.NoError(t, poolMock.ExpectationsWereMet())
	})
}

func TestGetSessionByUUID(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("TEST")
	queries := &SessionQueries{}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	repo := New(log, poolMock)
	query := regexp.QuoteMeta(queries.GetSessionByUUID())
	session := testSession()

	t.Run("success", func(t *testing.T) {
		rows := pgxmock.NewRows(sessionColumns()).
			AddRow(
				session.UUID, session.UserID, session.RefreshToken,
				session.LastUsedAt, session.OS, session.OSVersion,
				session.Browser, session.BrowserVersion, session.IPv4Address,
				session.CreatedAt, session.ExpiresAt,
			)
		poolMock.ExpectQuery(query).WithArgs("sess-uuid-123").WillReturnRows(rows)

		result, err := repo.GetSessionByUUID(ctx, "sess-uuid-123")
		require.NoError(t, err)
		assert.Equal(t, session.UUID, result.UUID)
		assert.Equal(t, session.UserID, result.UserID)
		assert.Equal(t, session.RefreshToken, result.RefreshToken)
		assert.Equal(t, session.OS, result.OS)
		assert.Equal(t, session.Browser, result.Browser)
		assert.Equal(t, session.IPv4Address, result.IPv4Address)
		assert.NoError(t, poolMock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		poolMock.ExpectQuery(query).WithArgs("nonexistent").WillReturnError(pgx.ErrNoRows)

		result, err := repo.GetSessionByUUID(ctx, "nonexistent")
		assert.Nil(t, result)
		require.Error(t, err)

		var appErr *errs.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, errs.ErrorNotFound, appErr.Code)
		assert.NoError(t, poolMock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		poolMock.ExpectQuery(query).WithArgs("sess-uuid-123").WillReturnError(errors.New("failed to connect to host"))

		result, err := repo.GetSessionByUUID(ctx, "sess-uuid-123")
		assert.Nil(t, result)
		require.Error(t, err)
		assert.NoError(t, poolMock.ExpectationsWereMet())
	})
}

func TestDeleteSession(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("TEST")
	queries := &SessionQueries{}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	repo := New(log, poolMock)
	query := regexp.QuoteMeta(queries.DeleteSession())

	t.Run("success", func(t *testing.T) {
		poolMock.ExpectExec(query).WithArgs("sess-uuid-123").
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.DeleteSession(ctx, "sess-uuid-123")
		assert.NoError(t, err)
		assert.NoError(t, poolMock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		poolMock.ExpectExec(query).WithArgs("nonexistent").
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err := repo.DeleteSession(ctx, "nonexistent")
		require.Error(t, err)

		var appErr *errs.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, errs.ErrorNotFound, appErr.Code)
		assert.NoError(t, poolMock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		poolMock.ExpectExec(query).WithArgs("sess-uuid-123").
			WillReturnError(errors.New("failed to connect to host"))

		err := repo.DeleteSession(ctx, "sess-uuid-123")
		require.Error(t, err)
		assert.NoError(t, poolMock.ExpectationsWereMet())
	})
}

func TestDeleteAllUserSessions(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("TEST")
	queries := &SessionQueries{}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	repo := New(log, poolMock)
	query := regexp.QuoteMeta(queries.DeleteAllUserSessions())

	t.Run("success", func(t *testing.T) {
		poolMock.ExpectExec(query).WithArgs("user-uuid-456").
			WillReturnResult(pgxmock.NewResult("DELETE", 3))

		err := repo.DeleteAllUserSessions(ctx, "user-uuid-456")
		assert.NoError(t, err)
		assert.NoError(t, poolMock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		poolMock.ExpectExec(query).WithArgs("user-uuid-456").
			WillReturnError(errors.New("failed to connect to host"))

		err := repo.DeleteAllUserSessions(ctx, "user-uuid-456")
		require.Error(t, err)
		assert.NoError(t, poolMock.ExpectationsWereMet())
	})
}

func TestUpdateSessionLastUsed(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("TEST")
	queries := &SessionQueries{}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer poolMock.Close()

	repo := New(log, poolMock)
	query := regexp.QuoteMeta(queries.UpdateSessionLastUsed())

	t.Run("success", func(t *testing.T) {
		poolMock.ExpectExec(query).WithArgs("sess-uuid-123").
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.UpdateSessionLastUsed(ctx, "sess-uuid-123")
		assert.NoError(t, err)
		assert.NoError(t, poolMock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		poolMock.ExpectExec(query).WithArgs("nonexistent").
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		err := repo.UpdateSessionLastUsed(ctx, "nonexistent")
		require.Error(t, err)

		var appErr *errs.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, errs.ErrorNotFound, appErr.Code)
		assert.NoError(t, poolMock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		poolMock.ExpectExec(query).WithArgs("sess-uuid-123").
			WillReturnError(errors.New("failed to connect to host"))

		err := repo.UpdateSessionLastUsed(ctx, "sess-uuid-123")
		require.Error(t, err)
		assert.NoError(t, poolMock.ExpectationsWereMet())
	})
}
