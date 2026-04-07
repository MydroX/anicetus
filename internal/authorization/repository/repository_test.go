package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"MydroX/anicetus/internal/authorization/models"
	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	roleCols = []string{"uuid", "name", "description", "created_at", "updated_at"}
	permCols = []string{"uuid", "name", "description", "created_at"}
)

func setupTest(t *testing.T) (context.Context, AuthorizationRepository, pgxmock.PgxPoolIface, *AuthorizationQueries) {
	t.Helper()
	ctx := context.Background()
	log, err := logger.New("TEST")
	if err != nil {
		panic(err)
	}

	poolMock, err := pgxmock.NewPool()
	require.NoError(t, err, "Failed to create mock pool")

	repo := New(log, poolMock)
	queries := &AuthorizationQueries{}
	return ctx, repo, poolMock, queries
}

// --- Roles ---

func TestCreateRole(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.CreateRole())
	testRole := &models.Role{UUID: "role-uuid-1", Name: "admin", Description: "Administrator"}

	tests := []struct {
		name      string
		mockSetup func()
		wantCode  int
	}{
		{
			name: "Success",
			mockSetup: func() {
				poolMock.ExpectExec(query).
					WithArgs(testRole.UUID, testRole.Name, testRole.Description).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			name: "Unique violation",
			mockSetup: func() {
				poolMock.ExpectExec(query).
					WithArgs(testRole.UUID, testRole.Name, testRole.Description).
					WillReturnError(&pgconn.PgError{Code: "23505", Message: "duplicate key"})
			},
			wantCode: errs.ErrorUniqueViolation,
		},
		{
			name: "Database error",
			mockSetup: func() {
				poolMock.ExpectExec(query).
					WithArgs(testRole.UUID, testRole.Name, testRole.Description).
					WillReturnError(&pgconn.PgError{Code: "57P03", Message: "connection error"})
			},
			wantCode: errs.ErrorDatabaseUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()
			err := repo.CreateRole(ctx, testRole)

			if tc.wantCode == 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr)) {
					assert.Equal(t, tc.wantCode, appErr.Code)
				}
			}
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestGetRoleByUUID(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.GetRoleByUUID())
	testUUID := "role-uuid-1"
	now := time.Now()

	tests := map[string]struct {
		mockSetup   func()
		checkResult func(t *testing.T, role *models.Role, err error)
	}{
		"Success": {
			mockSetup: func() {
				rows := pgxmock.NewRows(roleCols).
					AddRow(testUUID, "admin", "Administrator", now, now)
				poolMock.ExpectQuery(query).WithArgs(testUUID).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, role *models.Role, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, role)
				assert.Equal(t, testUUID, role.UUID)
				assert.Equal(t, "admin", role.Name)
				assert.Equal(t, "Administrator", role.Description)
			},
		},
		"Not Found": {
			mockSetup: func() {
				poolMock.ExpectQuery(query).WithArgs(testUUID).WillReturnError(pgx.ErrNoRows)
			},
			checkResult: func(t *testing.T, role *models.Role, err error) {
				assert.Error(t, err)
				assert.Nil(t, role)
				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr)) {
					assert.Equal(t, errs.ErrorNotFound, appErr.Code)
				}
			},
		},
		"Database error": {
			mockSetup: func() {
				poolMock.ExpectQuery(query).WithArgs(testUUID).
					WillReturnError(&pgconn.PgError{Code: "57P03", Message: "connection error"})
			},
			checkResult: func(t *testing.T, role *models.Role, err error) {
				assert.Error(t, err)
				assert.Nil(t, role)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.mockSetup()
			role, err := repo.GetRoleByUUID(ctx, testUUID)
			tc.checkResult(t, role, err)
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestGetRoleByName(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.GetRoleByName())
	testName := "admin"
	now := time.Now()

	tests := map[string]struct {
		mockSetup   func()
		checkResult func(t *testing.T, role *models.Role, err error)
	}{
		"Success": {
			mockSetup: func() {
				rows := pgxmock.NewRows(roleCols).
					AddRow("role-uuid-1", testName, "Administrator", now, now)
				poolMock.ExpectQuery(query).WithArgs(testName).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, role *models.Role, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, role)
				assert.Equal(t, testName, role.Name)
				assert.Equal(t, "role-uuid-1", role.UUID)
			},
		},
		"Not Found": {
			mockSetup: func() {
				poolMock.ExpectQuery(query).WithArgs(testName).WillReturnError(pgx.ErrNoRows)
			},
			checkResult: func(t *testing.T, role *models.Role, err error) {
				assert.Error(t, err)
				assert.Nil(t, role)
				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr)) {
					assert.Equal(t, errs.ErrorNotFound, appErr.Code)
				}
			},
		},
		"Database error": {
			mockSetup: func() {
				poolMock.ExpectQuery(query).WithArgs(testName).
					WillReturnError(&pgconn.PgError{Code: "57P03", Message: "connection error"})
			},
			checkResult: func(t *testing.T, role *models.Role, err error) {
				assert.Error(t, err)
				assert.Nil(t, role)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.mockSetup()
			role, err := repo.GetRoleByName(ctx, testName)
			tc.checkResult(t, role, err)
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestGetAllRoles(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.GetAllRoles())
	now := time.Now()

	tests := map[string]struct {
		mockSetup   func()
		checkResult func(t *testing.T, roles []*models.Role, err error)
	}{
		"Success Multiple Roles": {
			mockSetup: func() {
				rows := pgxmock.NewRows(roleCols).
					AddRow("role-uuid-1", "admin", "Administrator", now, now).
					AddRow("role-uuid-2", "editor", "Editor", now, now)
				poolMock.ExpectQuery(query).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, roles []*models.Role, err error) {
				assert.NoError(t, err)
				assert.Len(t, roles, 2)
				assert.Equal(t, "admin", roles[0].Name)
				assert.Equal(t, "editor", roles[1].Name)
			},
		},
		"Success Empty Result": {
			mockSetup: func() {
				rows := pgxmock.NewRows(roleCols)
				poolMock.ExpectQuery(query).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, roles []*models.Role, err error) {
				assert.NoError(t, err)
				assert.Empty(t, roles)
			},
		},
		"Query Error": {
			mockSetup: func() {
				poolMock.ExpectQuery(query).WillReturnError(pgx.ErrNoRows)
			},
			checkResult: func(t *testing.T, roles []*models.Role, err error) {
				assert.Error(t, err)
				assert.Nil(t, roles)
			},
		},
		"Scan Error": {
			mockSetup: func() {
				rows := pgxmock.NewRows(roleCols).
					AddRow("role-uuid-1", "admin", "Administrator", now, now).
					AddRow("role-uuid-2", "editor", "Editor", now, now).
					RowError(1, errors.New("scan error"))
				poolMock.ExpectQuery(query).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, roles []*models.Role, err error) {
				assert.Error(t, err)
				assert.Nil(t, roles)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.mockSetup()
			roles, err := repo.GetAllRoles(ctx)
			tc.checkResult(t, roles, err)
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestUpdateRole(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.UpdateRole())
	testRole := &models.Role{UUID: "role-uuid-1", Name: "admin-updated", Description: "Updated"}

	tests := []struct {
		name     string
		mockSetup func()
		wantCode int
	}{
		{
			name: "Success",
			mockSetup: func() {
				poolMock.ExpectExec(query).
					WithArgs(testRole.Name, testRole.Description, testRole.UUID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			name: "Not Found",
			mockSetup: func() {
				poolMock.ExpectExec(query).
					WithArgs(testRole.Name, testRole.Description, testRole.UUID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			wantCode: errs.ErrorNotFound,
		},
		{
			name: "Database error",
			mockSetup: func() {
				poolMock.ExpectExec(query).
					WithArgs(testRole.Name, testRole.Description, testRole.UUID).
					WillReturnError(&pgconn.PgError{Code: "57P03", Message: "connection error"})
			},
			wantCode: errs.ErrorDatabaseUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()
			err := repo.UpdateRole(ctx, testRole)

			if tc.wantCode == 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr)) {
					assert.Equal(t, tc.wantCode, appErr.Code)
				}
			}
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestDeleteRole(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.DeleteRole())
	testUUID := "role-uuid-1"

	tests := []struct {
		name     string
		mockSetup func()
		wantCode int
	}{
		{
			name: "Success",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(testUUID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
		},
		{
			name: "Not Found",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(testUUID).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			wantCode: errs.ErrorNotFound,
		},
		{
			name: "Database error",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(testUUID).
					WillReturnError(&pgconn.PgError{Code: "57P03", Message: "connection error"})
			},
			wantCode: errs.ErrorDatabaseUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()
			err := repo.DeleteRole(ctx, testUUID)

			if tc.wantCode == 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr)) {
					assert.Equal(t, tc.wantCode, appErr.Code)
				}
			}
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

// --- Permissions ---

func TestCreatePermission(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.CreatePermission())
	testPerm := &models.Permission{UUID: "perm-uuid-1", Name: "users:read", Description: "Read users"}

	tests := []struct {
		name     string
		mockSetup func()
		wantCode int
	}{
		{
			name: "Success",
			mockSetup: func() {
				poolMock.ExpectExec(query).
					WithArgs(testPerm.UUID, testPerm.Name, testPerm.Description).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			name: "Unique violation",
			mockSetup: func() {
				poolMock.ExpectExec(query).
					WithArgs(testPerm.UUID, testPerm.Name, testPerm.Description).
					WillReturnError(&pgconn.PgError{Code: "23505", Message: "duplicate key"})
			},
			wantCode: errs.ErrorUniqueViolation,
		},
		{
			name: "Database error",
			mockSetup: func() {
				poolMock.ExpectExec(query).
					WithArgs(testPerm.UUID, testPerm.Name, testPerm.Description).
					WillReturnError(&pgconn.PgError{Code: "57P03", Message: "connection error"})
			},
			wantCode: errs.ErrorDatabaseUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()
			err := repo.CreatePermission(ctx, testPerm)

			if tc.wantCode == 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr)) {
					assert.Equal(t, tc.wantCode, appErr.Code)
				}
			}
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestGetAllPermissions(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.GetAllPermissions())
	now := time.Now()

	tests := map[string]struct {
		mockSetup   func()
		checkResult func(t *testing.T, perms []*models.Permission, err error)
	}{
		"Success Multiple Permissions": {
			mockSetup: func() {
				rows := pgxmock.NewRows(permCols).
					AddRow("perm-uuid-1", "users:read", "Read users", now).
					AddRow("perm-uuid-2", "users:write", "Write users", now)
				poolMock.ExpectQuery(query).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, perms []*models.Permission, err error) {
				assert.NoError(t, err)
				assert.Len(t, perms, 2)
				assert.Equal(t, "users:read", perms[0].Name)
				assert.Equal(t, "users:write", perms[1].Name)
			},
		},
		"Success Empty Result": {
			mockSetup: func() {
				rows := pgxmock.NewRows(permCols)
				poolMock.ExpectQuery(query).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, perms []*models.Permission, err error) {
				assert.NoError(t, err)
				assert.Empty(t, perms)
			},
		},
		"Query Error": {
			mockSetup: func() {
				poolMock.ExpectQuery(query).WillReturnError(pgx.ErrNoRows)
			},
			checkResult: func(t *testing.T, perms []*models.Permission, err error) {
				assert.Error(t, err)
				assert.Nil(t, perms)
			},
		},
		"Scan Error": {
			mockSetup: func() {
				rows := pgxmock.NewRows(permCols).
					AddRow("perm-uuid-1", "users:read", "Read users", now).
					AddRow("perm-uuid-2", "users:write", "Write users", now).
					RowError(1, errors.New("scan error"))
				poolMock.ExpectQuery(query).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, perms []*models.Permission, err error) {
				assert.Error(t, err)
				assert.Nil(t, perms)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.mockSetup()
			perms, err := repo.GetAllPermissions(ctx)
			tc.checkResult(t, perms, err)
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestDeletePermission(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.DeletePermission())
	testUUID := "perm-uuid-1"

	tests := []struct {
		name     string
		mockSetup func()
		wantCode int
	}{
		{
			name: "Success",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(testUUID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
		},
		{
			name: "Not Found",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(testUUID).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			wantCode: errs.ErrorNotFound,
		},
		{
			name: "Database error",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(testUUID).
					WillReturnError(&pgconn.PgError{Code: "57P03", Message: "connection error"})
			},
			wantCode: errs.ErrorDatabaseUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()
			err := repo.DeletePermission(ctx, testUUID)

			if tc.wantCode == 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr)) {
					assert.Equal(t, tc.wantCode, appErr.Code)
				}
			}
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

// --- Role-Permission assignments ---

func TestAssignPermissionToRole(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.AssignPermissionToRole())
	roleUUID := "role-uuid-1"
	permUUID := "perm-uuid-1"

	tests := []struct {
		name     string
		mockSetup func()
		wantCode int
	}{
		{
			name: "Success",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(roleUUID, permUUID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			name: "Unique violation",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(roleUUID, permUUID).
					WillReturnError(&pgconn.PgError{Code: "23505", Message: "duplicate key"})
			},
			wantCode: errs.ErrorUniqueViolation,
		},
		{
			name: "Database error",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(roleUUID, permUUID).
					WillReturnError(&pgconn.PgError{Code: "57P03", Message: "connection error"})
			},
			wantCode: errs.ErrorDatabaseUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()
			err := repo.AssignPermissionToRole(ctx, roleUUID, permUUID)

			if tc.wantCode == 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr)) {
					assert.Equal(t, tc.wantCode, appErr.Code)
				}
			}
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestRemovePermissionFromRole(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.RemovePermissionFromRole())
	roleUUID := "role-uuid-1"
	permUUID := "perm-uuid-1"

	tests := []struct {
		name     string
		mockSetup func()
		wantCode int
	}{
		{
			name: "Success",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(roleUUID, permUUID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
		},
		{
			name: "Not Found",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(roleUUID, permUUID).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			wantCode: errs.ErrorNotFound,
		},
		{
			name: "Database error",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(roleUUID, permUUID).
					WillReturnError(&pgconn.PgError{Code: "57P03", Message: "connection error"})
			},
			wantCode: errs.ErrorDatabaseUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()
			err := repo.RemovePermissionFromRole(ctx, roleUUID, permUUID)

			if tc.wantCode == 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr)) {
					assert.Equal(t, tc.wantCode, appErr.Code)
				}
			}
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestGetRolePermissions(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.GetRolePermissions())
	roleUUID := "role-uuid-1"
	now := time.Now()

	tests := map[string]struct {
		mockSetup   func()
		checkResult func(t *testing.T, perms []*models.Permission, err error)
	}{
		"Success Multiple Permissions": {
			mockSetup: func() {
				rows := pgxmock.NewRows(permCols).
					AddRow("perm-uuid-1", "users:read", "Read users", now).
					AddRow("perm-uuid-2", "users:write", "Write users", now)
				poolMock.ExpectQuery(query).WithArgs(roleUUID).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, perms []*models.Permission, err error) {
				assert.NoError(t, err)
				assert.Len(t, perms, 2)
				assert.Equal(t, "users:read", perms[0].Name)
				assert.Equal(t, "users:write", perms[1].Name)
			},
		},
		"Success Empty Result": {
			mockSetup: func() {
				rows := pgxmock.NewRows(permCols)
				poolMock.ExpectQuery(query).WithArgs(roleUUID).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, perms []*models.Permission, err error) {
				assert.NoError(t, err)
				assert.Empty(t, perms)
			},
		},
		"Query Error": {
			mockSetup: func() {
				poolMock.ExpectQuery(query).WithArgs(roleUUID).WillReturnError(pgx.ErrNoRows)
			},
			checkResult: func(t *testing.T, perms []*models.Permission, err error) {
				assert.Error(t, err)
				assert.Nil(t, perms)
			},
		},
		"Scan Error": {
			mockSetup: func() {
				rows := pgxmock.NewRows(permCols).
					AddRow("perm-uuid-1", "users:read", "Read users", now).
					AddRow("perm-uuid-2", "users:write", "Write users", now).
					RowError(1, errors.New("scan error"))
				poolMock.ExpectQuery(query).WithArgs(roleUUID).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, perms []*models.Permission, err error) {
				assert.Error(t, err)
				assert.Nil(t, perms)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.mockSetup()
			perms, err := repo.GetRolePermissions(ctx, roleUUID)
			tc.checkResult(t, perms, err)
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

// --- User-Role assignments ---

func TestAssignRoleToUser(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.AssignRoleToUser())
	userUUID := "user-uuid-1"
	roleUUID := "role-uuid-1"

	tests := []struct {
		name     string
		mockSetup func()
		wantCode int
	}{
		{
			name: "Success",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(userUUID, roleUUID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			name: "Unique violation",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(userUUID, roleUUID).
					WillReturnError(&pgconn.PgError{Code: "23505", Message: "duplicate key"})
			},
			wantCode: errs.ErrorUniqueViolation,
		},
		{
			name: "Database error",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(userUUID, roleUUID).
					WillReturnError(&pgconn.PgError{Code: "57P03", Message: "connection error"})
			},
			wantCode: errs.ErrorDatabaseUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()
			err := repo.AssignRoleToUser(ctx, userUUID, roleUUID)

			if tc.wantCode == 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr)) {
					assert.Equal(t, tc.wantCode, appErr.Code)
				}
			}
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestRemoveRoleFromUser(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.RemoveRoleFromUser())
	userUUID := "user-uuid-1"
	roleUUID := "role-uuid-1"

	tests := []struct {
		name     string
		mockSetup func()
		wantCode int
	}{
		{
			name: "Success",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(userUUID, roleUUID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
		},
		{
			name: "Not Found",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(userUUID, roleUUID).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			wantCode: errs.ErrorNotFound,
		},
		{
			name: "Database error",
			mockSetup: func() {
				poolMock.ExpectExec(query).WithArgs(userUUID, roleUUID).
					WillReturnError(&pgconn.PgError{Code: "57P03", Message: "connection error"})
			},
			wantCode: errs.ErrorDatabaseUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()
			err := repo.RemoveRoleFromUser(ctx, userUUID, roleUUID)

			if tc.wantCode == 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *errs.AppError
				if assert.True(t, errors.As(err, &appErr)) {
					assert.Equal(t, tc.wantCode, appErr.Code)
				}
			}
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestGetUserRoles(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.GetUserRoles())
	userUUID := "user-uuid-1"
	now := time.Now()

	tests := map[string]struct {
		mockSetup   func()
		checkResult func(t *testing.T, roles []*models.Role, err error)
	}{
		"Success Multiple Roles": {
			mockSetup: func() {
				rows := pgxmock.NewRows(roleCols).
					AddRow("role-uuid-1", "admin", "Administrator", now, now).
					AddRow("role-uuid-2", "editor", "Editor", now, now)
				poolMock.ExpectQuery(query).WithArgs(userUUID).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, roles []*models.Role, err error) {
				assert.NoError(t, err)
				assert.Len(t, roles, 2)
				assert.Equal(t, "admin", roles[0].Name)
				assert.Equal(t, "editor", roles[1].Name)
			},
		},
		"Success Empty Result": {
			mockSetup: func() {
				rows := pgxmock.NewRows(roleCols)
				poolMock.ExpectQuery(query).WithArgs(userUUID).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, roles []*models.Role, err error) {
				assert.NoError(t, err)
				assert.Empty(t, roles)
			},
		},
		"Query Error": {
			mockSetup: func() {
				poolMock.ExpectQuery(query).WithArgs(userUUID).WillReturnError(pgx.ErrNoRows)
			},
			checkResult: func(t *testing.T, roles []*models.Role, err error) {
				assert.Error(t, err)
				assert.Nil(t, roles)
			},
		},
		"Scan Error": {
			mockSetup: func() {
				rows := pgxmock.NewRows(roleCols).
					AddRow("role-uuid-1", "admin", "Administrator", now, now).
					AddRow("role-uuid-2", "editor", "Editor", now, now).
					RowError(1, errors.New("scan error"))
				poolMock.ExpectQuery(query).WithArgs(userUUID).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, roles []*models.Role, err error) {
				assert.Error(t, err)
				assert.Nil(t, roles)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.mockSetup()
			roles, err := repo.GetUserRoles(ctx, userUUID)
			tc.checkResult(t, roles, err)
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}

func TestGetUserPermissions(t *testing.T) {
	ctx, repo, poolMock, queries := setupTest(t)
	defer poolMock.Close()

	query := regexp.QuoteMeta(queries.GetUserPermissions())
	userUUID := "user-uuid-1"

	tests := map[string]struct {
		mockSetup   func()
		checkResult func(t *testing.T, perms []string, err error)
	}{
		"Success Multiple Permissions": {
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"name"}).
					AddRow("users:read").
					AddRow("users:write")
				poolMock.ExpectQuery(query).WithArgs(userUUID).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, perms []string, err error) {
				assert.NoError(t, err)
				assert.Len(t, perms, 2)
				assert.Equal(t, "users:read", perms[0])
				assert.Equal(t, "users:write", perms[1])
			},
		},
		"Success Empty Result": {
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"name"})
				poolMock.ExpectQuery(query).WithArgs(userUUID).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, perms []string, err error) {
				assert.NoError(t, err)
				assert.Empty(t, perms)
			},
		},
		"Query Error": {
			mockSetup: func() {
				poolMock.ExpectQuery(query).WithArgs(userUUID).WillReturnError(pgx.ErrNoRows)
			},
			checkResult: func(t *testing.T, perms []string, err error) {
				assert.Error(t, err)
				assert.Nil(t, perms)
			},
		},
		"Scan Error": {
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"name"}).
					AddRow("users:read").
					AddRow("users:write").
					RowError(1, errors.New("scan error"))
				poolMock.ExpectQuery(query).WithArgs(userUUID).WillReturnRows(rows)
			},
			checkResult: func(t *testing.T, perms []string, err error) {
				assert.Error(t, err)
				assert.Nil(t, perms)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.mockSetup()
			perms, err := repo.GetUserPermissions(ctx, userUUID)
			tc.checkResult(t, perms, err)
		})
	}

	assert.NoError(t, poolMock.ExpectationsWereMet())
}
