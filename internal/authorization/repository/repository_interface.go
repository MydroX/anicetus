package repository

import (
	"context"

	"MydroX/anicetus/internal/authorization/models"
)

//go:generate mockgen -source=repository_interface.go -destination=../mocks/mock_repository.go -package=mocks

type AuthorizationRepository interface {
	// Roles
	CreateRole(ctx context.Context, role *models.Role) error
	GetRoleByUUID(ctx context.Context, uuid string) (*models.Role, error)
	GetRoleByName(ctx context.Context, name string) (*models.Role, error)
	GetAllRoles(ctx context.Context) ([]*models.Role, error)
	UpdateRole(ctx context.Context, role *models.Role) error
	DeleteRole(ctx context.Context, uuid string) error

	// Permissions
	CreatePermission(ctx context.Context, perm *models.Permission) error
	GetAllPermissions(ctx context.Context) ([]*models.Permission, error)
	DeletePermission(ctx context.Context, uuid string) error

	// Role-Permission assignments
	AssignPermissionToRole(ctx context.Context, roleUUID, permissionUUID string) error
	RemovePermissionFromRole(ctx context.Context, roleUUID, permissionUUID string) error
	GetRolePermissions(ctx context.Context, roleUUID string) ([]*models.Permission, error)

	// User-Role assignments
	AssignRoleToUser(ctx context.Context, userUUID, roleUUID string) error
	RemoveRoleFromUser(ctx context.Context, userUUID, roleUUID string) error
	GetUserRoles(ctx context.Context, userUUID string) ([]*models.Role, error)
	GetUserPermissions(ctx context.Context, userUUID string) ([]string, error)
}
