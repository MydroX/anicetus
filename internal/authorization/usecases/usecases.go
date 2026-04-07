package usecases

import (
	"context"

	"MydroX/anicetus/internal/authorization/dto"
	"MydroX/anicetus/internal/authorization/models"
	"MydroX/anicetus/internal/authorization/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type usecases struct {
	logger     *zap.SugaredLogger
	repository repository.AuthorizationRepository
}

func New(l *zap.SugaredLogger, r repository.AuthorizationRepository) AuthorizationUsecases {
	return &usecases{
		logger:     l,
		repository: r,
	}
}

// Roles

func (u *usecases) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) error {
	role := models.Role{
		UUID:        uuid.Must(uuid.NewV7()).String(),
		Name:        req.Name,
		Description: req.Description,
	}

	return u.repository.CreateRole(ctx, &role)
}

func (u *usecases) GetAllRoles(ctx context.Context) ([]*dto.RoleResponse, error) {
	roles, err := u.repository.GetAllRoles(ctx)
	if err != nil {
		return nil, err
	}

	var result []*dto.RoleResponse
	for _, role := range roles {
		result = append(result, &dto.RoleResponse{
			UUID:        role.UUID,
			Name:        role.Name,
			Description: role.Description,
		})
	}

	return result, nil
}

func (u *usecases) UpdateRole(ctx context.Context, roleUUID string, req *dto.UpdateRoleRequest) error {
	role := models.Role{
		UUID:        roleUUID,
		Name:        req.Name,
		Description: req.Description,
	}

	return u.repository.UpdateRole(ctx, &role)
}

func (u *usecases) DeleteRole(ctx context.Context, uuid string) error {
	return u.repository.DeleteRole(ctx, uuid)
}

// Permissions

func (u *usecases) CreatePermission(ctx context.Context, req *dto.CreatePermissionRequest) error {
	perm := models.Permission{
		UUID:        uuid.Must(uuid.NewV7()).String(),
		Name:        req.Name,
		Description: req.Description,
	}

	return u.repository.CreatePermission(ctx, &perm)
}

func (u *usecases) GetAllPermissions(ctx context.Context) ([]*dto.PermissionResponse, error) {
	permissions, err := u.repository.GetAllPermissions(ctx)
	if err != nil {
		return nil, err
	}

	var result []*dto.PermissionResponse
	for _, perm := range permissions {
		result = append(result, &dto.PermissionResponse{
			UUID:        perm.UUID,
			Name:        perm.Name,
			Description: perm.Description,
		})
	}

	return result, nil
}

// Assignments

func (u *usecases) AssignPermissionToRole(ctx context.Context, roleUUID, permissionUUID string) error {
	return u.repository.AssignPermissionToRole(ctx, roleUUID, permissionUUID)
}

func (u *usecases) RemovePermissionFromRole(ctx context.Context, roleUUID, permissionUUID string) error {
	return u.repository.RemovePermissionFromRole(ctx, roleUUID, permissionUUID)
}

func (u *usecases) AssignRoleToUser(ctx context.Context, userUUID, roleUUID string) error {
	return u.repository.AssignRoleToUser(ctx, userUUID, roleUUID)
}

func (u *usecases) RemoveRoleFromUser(ctx context.Context, userUUID, roleUUID string) error {
	return u.repository.RemoveRoleFromUser(ctx, userUUID, roleUUID)
}

func (u *usecases) GetUserRoles(ctx context.Context, userUUID string) ([]*dto.RoleResponse, error) {
	roles, err := u.repository.GetUserRoles(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	var result []*dto.RoleResponse
	for _, role := range roles {
		result = append(result, &dto.RoleResponse{
			UUID:        role.UUID,
			Name:        role.Name,
			Description: role.Description,
		})
	}

	return result, nil
}

func (u *usecases) GetUserPermissions(ctx context.Context, userUUID string) ([]string, error) {
	return u.repository.GetUserPermissions(ctx, userUUID)
}
