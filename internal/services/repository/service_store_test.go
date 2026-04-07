package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (ServiceStore, pgxmock.PgxPoolIface, *ServiceQueries) {
	log, _ := logger.New("TEST")
	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	t.Cleanup(func() { poolMock.Close() })
	return New(log, poolMock), poolMock, &ServiceQueries{}
}

func TestIsValidService(t *testing.T) {
	repo, pool, q := setupTest(t)
	ctx := context.Background()
	query := regexp.QuoteMeta(q.IsValidService())

	t.Run("exists", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"exists"}).AddRow(true)
		pool.ExpectQuery(query).WithArgs("my-service").WillReturnRows(rows)

		exists, err := repo.IsValidService(ctx, "my-service")
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("not exists", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"exists"}).AddRow(false)
		pool.ExpectQuery(query).WithArgs("unknown").WillReturnRows(rows)

		exists, err := repo.IsValidService(ctx, "unknown")
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		pool.ExpectQuery(query).WithArgs("svc").WillReturnError(errors.New("failed to connect to host"))

		_, err := repo.IsValidService(ctx, "svc")
		assert.Error(t, err)
		assert.NoError(t, pool.ExpectationsWereMet())
	})
}

func TestGetAllowedServices(t *testing.T) {
	repo, pool, q := setupTest(t)
	ctx := context.Background()
	query := regexp.QuoteMeta(q.GetAllowedServices())

	t.Run("success", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"audience"}).
			AddRow("service-a").
			AddRow("service-b")
		pool.ExpectQuery(query).WillReturnRows(rows)

		services, err := repo.GetAllowedServices(ctx)
		require.NoError(t, err)
		assert.Equal(t, []string{"service-a", "service-b"}, services)
		assert.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("empty", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"audience"})
		pool.ExpectQuery(query).WillReturnRows(rows)

		services, err := repo.GetAllowedServices(ctx)
		require.NoError(t, err)
		assert.Nil(t, services)
		assert.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		pool.ExpectQuery(query).WillReturnError(errors.New("failed to connect to host"))

		_, err := repo.GetAllowedServices(ctx)
		assert.Error(t, err)
		assert.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"audience"}).
			AddRow("service-a").
			RowError(0, errors.New("scan failed"))
		pool.ExpectQuery(query).WillReturnRows(rows)

		_, err := repo.GetAllowedServices(ctx)
		assert.Error(t, err)
		assert.NoError(t, pool.ExpectationsWereMet())
	})
}

func TestGetUserServices(t *testing.T) {
	repo, pool, q := setupTest(t)
	ctx := context.Background()
	query := regexp.QuoteMeta(q.GetUserServices())

	t.Run("success", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"audience"}).AddRow("svc-1")
		pool.ExpectQuery(query).WithArgs("user-uuid").WillReturnRows(rows)

		services, err := repo.GetUserServices(ctx, "user-uuid")
		require.NoError(t, err)
		assert.Equal(t, []string{"svc-1"}, services)
		assert.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("empty", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"audience"})
		pool.ExpectQuery(query).WithArgs("user-uuid").WillReturnRows(rows)

		services, err := repo.GetUserServices(ctx, "user-uuid")
		require.NoError(t, err)
		assert.Nil(t, services)
		assert.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		pool.ExpectQuery(query).WithArgs("user-uuid").WillReturnError(pgx.ErrNoRows)

		_, err := repo.GetUserServices(ctx, "user-uuid")
		assert.Error(t, err)
		assert.NoError(t, pool.ExpectationsWereMet())
	})
}

func TestRegisterService(t *testing.T) {
	repo, pool, q := setupTest(t)
	ctx := context.Background()
	query := regexp.QuoteMeta(q.RegisterService())

	t.Run("success", func(t *testing.T) {
		pool.ExpectExec(query).
			WithArgs(pgxmock.AnyArg(), "my-audience", "My Service", "A description", pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.RegisterService(ctx, "my-audience", map[string]any{
			"service_name": "My Service",
			"description":  "A description",
			"permissions":  map[string]any{"read": true},
		})
		assert.NoError(t, err)
		assert.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("missing metadata", func(t *testing.T) {
		pool.ExpectExec(query).
			WithArgs(pgxmock.AnyArg(), "my-audience", "", "", pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.RegisterService(ctx, "my-audience", map[string]any{})
		assert.NoError(t, err)
		assert.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("unique violation", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23505"}
		pool.ExpectExec(query).
			WithArgs(pgxmock.AnyArg(), "dup-svc", "", "", pgxmock.AnyArg()).
			WillReturnError(pgErr)

		err := repo.RegisterService(ctx, "dup-svc", map[string]any{})
		require.Error(t, err)

		var appErr *errs.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, errs.ErrorUniqueViolation, appErr.Code)
		assert.NoError(t, pool.ExpectationsWereMet())
	})
}

func TestRevokeService(t *testing.T) {
	repo, pool, q := setupTest(t)
	ctx := context.Background()
	query := regexp.QuoteMeta(q.RevokeService())

	t.Run("success", func(t *testing.T) {
		pool.ExpectExec(query).WithArgs("my-svc").
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		assert.NoError(t, repo.RevokeService(ctx, "my-svc"))
		assert.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		pool.ExpectExec(query).WithArgs("unknown").
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		err := repo.RevokeService(ctx, "unknown")
		require.Error(t, err)

		var appErr *errs.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, errs.ErrorNotFound, appErr.Code)
		assert.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		pool.ExpectExec(query).WithArgs("svc").
			WillReturnError(errors.New("failed to connect to host"))

		assert.Error(t, repo.RevokeService(ctx, "svc"))
		assert.NoError(t, pool.ExpectationsWereMet())
	})
}

func TestAssignServiceToUser(t *testing.T) {
	repo, pool, q := setupTest(t)
	ctx := context.Background()
	query := regexp.QuoteMeta(q.AssignServiceToUser())

	t.Run("success", func(t *testing.T) {
		pool.ExpectExec(query).WithArgs("user-uuid", "svc").
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		assert.NoError(t, repo.AssignServiceToUser(ctx, "user-uuid", "svc"))
		assert.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("not found or inactive", func(t *testing.T) {
		pool.ExpectExec(query).WithArgs("user-uuid", "inactive-svc").
			WillReturnResult(pgxmock.NewResult("INSERT", 0))

		err := repo.AssignServiceToUser(ctx, "user-uuid", "inactive-svc")
		require.Error(t, err)

		var appErr *errs.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, errs.ErrorNotFound, appErr.Code)
		assert.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("fk violation", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23503"}
		pool.ExpectExec(query).WithArgs("user-uuid", "svc").WillReturnError(pgErr)

		err := repo.AssignServiceToUser(ctx, "user-uuid", "svc")
		require.Error(t, err)

		var appErr *errs.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, errs.ErrorForeignKeyViolation, appErr.Code)
		assert.NoError(t, pool.ExpectationsWereMet())
	})
}

func TestUnassignServiceFromUser(t *testing.T) {
	repo, pool, q := setupTest(t)
	ctx := context.Background()
	query := regexp.QuoteMeta(q.UnassignServiceFromUser())

	t.Run("success", func(t *testing.T) {
		pool.ExpectExec(query).WithArgs("user-uuid", "svc").
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		assert.NoError(t, repo.UnassignServiceFromUser(ctx, "user-uuid", "svc"))
		assert.NoError(t, pool.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		pool.ExpectExec(query).WithArgs("user-uuid", "svc").
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err := repo.UnassignServiceFromUser(ctx, "user-uuid", "svc")
		require.Error(t, err)

		var appErr *errs.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, errs.ErrorNotFound, appErr.Code)
		assert.NoError(t, pool.ExpectationsWereMet())
	})
}
