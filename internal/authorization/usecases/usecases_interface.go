package usecases

import (
	"context"

	"MydroX/anicetus/internal/authorization/dto"
)

//go:generate mockgen -source=usecases_interface.go -destination=../mocks/mock_usecases.go -package=mocks

type AuthorizationUsecases interface {
	// Role CRUD
	CreateRole(ctx context.Context, req *dto.CreateRoleRequest) error
	GetAllRoles(ctx context.Context) ([]*dto.RoleResponse, error)
	UpdateRole(ctx context.Context, uuid string, req *dto.UpdateRoleRequest) error
	DeleteRole(ctx context.Context, uuid string) error

	// Permission CRUD
	CreatePermission(ctx context.Context, req *dto.CreatePermissionRequest) error
	GetAllPermissions(ctx context.Context) ([]*dto.PermissionResponse, error)

	// Assignments
	AssignPermissionToRole(ctx context.Context, roleUUID, permissionUUID string) error
	RemovePermissionFromRole(ctx context.Context, roleUUID, permissionUUID string) error
	AssignRoleToUser(ctx context.Context, userUUID, roleUUID string) error
	RemoveRoleFromUser(ctx context.Context, userUUID, roleUUID string) error
	GetUserRoles(ctx context.Context, userUUID string) ([]*dto.RoleResponse, error)
	GetUserPermissions(ctx context.Context, userUUID string) ([]string, error)
}
