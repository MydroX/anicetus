package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"MydroX/anicetus/internal/authorization/dto"
	"MydroX/anicetus/internal/authorization/mocks"
	"MydroX/anicetus/internal/authorization/models"
	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupTest(t *testing.T) (*mocks.MockAuthorizationRepository, AuthorizationUsecases) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockAuthorizationRepository(ctrl)
	log, _ := logger.New("TEST")
	uc := New(log, repo)
	return repo, uc
}

// --- Roles ---

func TestCreateRole(t *testing.T) {
	repo, uc := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo.EXPECT().CreateRole(gomock.Any(), gomock.Any()).Return(nil)

		err := uc.CreateRole(ctx, &dto.CreateRoleRequest{Name: "admin", Description: "Admin role"})
		assert.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		repo.EXPECT().CreateRole(gomock.Any(), gomock.Any()).
			Return(errs.New(errs.ErrorUniqueViolation, "exists", nil))

		err := uc.CreateRole(ctx, &dto.CreateRoleRequest{Name: "admin"})
		require.Error(t, err)

		var appErr *errs.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, errs.ErrorUniqueViolation, appErr.Code)
	})
}

func TestGetAllRoles(t *testing.T) {
	repo, uc := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		repo.EXPECT().GetAllRoles(gomock.Any()).Return([]*models.Role{
			{UUID: "r1", Name: "admin", Description: "Admin", CreatedAt: now, UpdatedAt: now},
			{UUID: "r2", Name: "user", Description: "User", CreatedAt: now, UpdatedAt: now},
		}, nil)

		result, err := uc.GetAllRoles(ctx)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "admin", result[0].Name)
		assert.Equal(t, "r1", result[0].UUID)
	})

	t.Run("empty", func(t *testing.T) {
		repo.EXPECT().GetAllRoles(gomock.Any()).Return(nil, nil)

		result, err := uc.GetAllRoles(ctx)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("repo error", func(t *testing.T) {
		repo.EXPECT().GetAllRoles(gomock.Any()).Return(nil, errs.New(errs.ErrorInternal, "db error", nil))

		_, err := uc.GetAllRoles(ctx)
		assert.Error(t, err)
	})
}

func TestUpdateRole(t *testing.T) {
	repo, uc := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo.EXPECT().UpdateRole(gomock.Any(), gomock.Any()).Return(nil)

		err := uc.UpdateRole(ctx, "role-uuid", &dto.UpdateRoleRequest{Name: "updated"})
		assert.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		repo.EXPECT().UpdateRole(gomock.Any(), gomock.Any()).
			Return(errs.New(errs.ErrorNotFound, "not found", nil))

		err := uc.UpdateRole(ctx, "role-uuid", &dto.UpdateRoleRequest{Name: "updated"})
		assert.Error(t, err)
	})
}

func TestDeleteRole(t *testing.T) {
	repo, uc := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo.EXPECT().DeleteRole(gomock.Any(), "role-uuid").Return(nil)
		assert.NoError(t, uc.DeleteRole(ctx, "role-uuid"))
	})

	t.Run("repo error", func(t *testing.T) {
		repo.EXPECT().DeleteRole(gomock.Any(), "role-uuid").
			Return(errs.New(errs.ErrorNotFound, "not found", nil))
		assert.Error(t, uc.DeleteRole(ctx, "role-uuid"))
	})
}

// --- Permissions ---

func TestCreatePermission(t *testing.T) {
	repo, uc := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo.EXPECT().CreatePermission(gomock.Any(), gomock.Any()).Return(nil)

		err := uc.CreatePermission(ctx, &dto.CreatePermissionRequest{Name: "read:users"})
		assert.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		repo.EXPECT().CreatePermission(gomock.Any(), gomock.Any()).
			Return(errs.New(errs.ErrorUniqueViolation, "exists", nil))

		err := uc.CreatePermission(ctx, &dto.CreatePermissionRequest{Name: "read:users"})
		assert.Error(t, err)
	})
}

func TestGetAllPermissions(t *testing.T) {
	repo, uc := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		repo.EXPECT().GetAllPermissions(gomock.Any()).Return([]*models.Permission{
			{UUID: "p1", Name: "read:users", Description: "Read users", CreatedAt: now},
		}, nil)

		result, err := uc.GetAllPermissions(ctx)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "read:users", result[0].Name)
	})

	t.Run("repo error", func(t *testing.T) {
		repo.EXPECT().GetAllPermissions(gomock.Any()).Return(nil, errs.New(errs.ErrorInternal, "err", nil))

		_, err := uc.GetAllPermissions(ctx)
		assert.Error(t, err)
	})
}

// --- Assignments ---

func TestAssignPermissionToRole(t *testing.T) {
	repo, uc := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo.EXPECT().AssignPermissionToRole(gomock.Any(), "role-uuid", "perm-uuid").Return(nil)
		assert.NoError(t, uc.AssignPermissionToRole(ctx, "role-uuid", "perm-uuid"))
	})

	t.Run("repo error", func(t *testing.T) {
		repo.EXPECT().AssignPermissionToRole(gomock.Any(), "role-uuid", "perm-uuid").
			Return(errs.New(errs.ErrorForeignKeyViolation, "not found", nil))
		assert.Error(t, uc.AssignPermissionToRole(ctx, "role-uuid", "perm-uuid"))
	})
}

func TestRemovePermissionFromRole(t *testing.T) {
	repo, uc := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo.EXPECT().RemovePermissionFromRole(gomock.Any(), "role-uuid", "perm-uuid").Return(nil)
		assert.NoError(t, uc.RemovePermissionFromRole(ctx, "role-uuid", "perm-uuid"))
	})

	t.Run("repo error", func(t *testing.T) {
		repo.EXPECT().RemovePermissionFromRole(gomock.Any(), "role-uuid", "perm-uuid").
			Return(errs.New(errs.ErrorNotFound, "not found", nil))
		assert.Error(t, uc.RemovePermissionFromRole(ctx, "role-uuid", "perm-uuid"))
	})
}

func TestAssignRoleToUser(t *testing.T) {
	repo, uc := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo.EXPECT().AssignRoleToUser(gomock.Any(), "user-uuid", "role-uuid").Return(nil)
		assert.NoError(t, uc.AssignRoleToUser(ctx, "user-uuid", "role-uuid"))
	})

	t.Run("repo error", func(t *testing.T) {
		repo.EXPECT().AssignRoleToUser(gomock.Any(), "user-uuid", "role-uuid").
			Return(errs.New(errs.ErrorForeignKeyViolation, "not found", nil))
		assert.Error(t, uc.AssignRoleToUser(ctx, "user-uuid", "role-uuid"))
	})
}

func TestRemoveRoleFromUser(t *testing.T) {
	repo, uc := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo.EXPECT().RemoveRoleFromUser(gomock.Any(), "user-uuid", "role-uuid").Return(nil)
		assert.NoError(t, uc.RemoveRoleFromUser(ctx, "user-uuid", "role-uuid"))
	})

	t.Run("repo error", func(t *testing.T) {
		repo.EXPECT().RemoveRoleFromUser(gomock.Any(), "user-uuid", "role-uuid").
			Return(errs.New(errs.ErrorNotFound, "not found", nil))
		assert.Error(t, uc.RemoveRoleFromUser(ctx, "user-uuid", "role-uuid"))
	})
}

func TestGetUserRoles(t *testing.T) {
	repo, uc := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		repo.EXPECT().GetUserRoles(gomock.Any(), "user-uuid").Return([]*models.Role{
			{UUID: "r1", Name: "admin", Description: "Admin", CreatedAt: now, UpdatedAt: now},
		}, nil)

		result, err := uc.GetUserRoles(ctx, "user-uuid")
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "admin", result[0].Name)
	})

	t.Run("repo error", func(t *testing.T) {
		repo.EXPECT().GetUserRoles(gomock.Any(), "user-uuid").
			Return(nil, errs.New(errs.ErrorInternal, "err", nil))

		_, err := uc.GetUserRoles(ctx, "user-uuid")
		assert.Error(t, err)
	})
}

func TestGetUserPermissions(t *testing.T) {
	repo, uc := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo.EXPECT().GetUserPermissions(gomock.Any(), "user-uuid").Return([]string{"read", "write"}, nil)

		result, err := uc.GetUserPermissions(ctx, "user-uuid")
		require.NoError(t, err)
		assert.Equal(t, []string{"read", "write"}, result)
	})

	t.Run("repo error", func(t *testing.T) {
		repo.EXPECT().GetUserPermissions(gomock.Any(), "user-uuid").
			Return(nil, errs.New(errs.ErrorInternal, "err", nil))

		_, err := uc.GetUserPermissions(ctx, "user-uuid")
		assert.Error(t, err)
	})
}
