package repository

import (
	"context"
	"fmt"

	"MydroX/anicetus/internal/authorization/models"
	"MydroX/anicetus/pkg/db/postgresql/pgx"
	"MydroX/anicetus/pkg/errs"
	"go.uber.org/zap"
)

type repository struct {
	logger  *zap.SugaredLogger
	dbPool  pgx.DBPool
	queries *AuthorizationQueries
}

func New(l *zap.SugaredLogger, dbPool pgx.DBPool) AuthorizationRepository {
	return &repository{
		logger:  l,
		dbPool:  dbPool,
		queries: &AuthorizationQueries{},
	}
}

// Roles

func (r *repository) CreateRole(ctx context.Context, role *models.Role) error {
	_, err := r.dbPool.Exec(ctx, r.queries.CreateRole(), role.UUID, role.Name, role.Description)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	return nil
}

func (r *repository) GetRoleByUUID(ctx context.Context, uuid string) (*models.Role, error) {
	var role models.Role

	err := r.dbPool.QueryRow(ctx, r.queries.GetRoleByUUID(), uuid).
		Scan(&role.UUID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}

	return &role, nil
}

func (r *repository) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role

	err := r.dbPool.QueryRow(ctx, r.queries.GetRoleByName(), name).
		Scan(&role.UUID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}

	return &role, nil
}

func (r *repository) GetAllRoles(ctx context.Context) ([]*models.Role, error) {
	rows, err := r.dbPool.Query(ctx, r.queries.GetAllRoles())
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}

	var roles []*models.Role

	for rows.Next() {
		var role models.Role

		if err := rows.Scan(&role.UUID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, &errs.AppError{Code: errs.ErrorInternal, Err: fmt.Errorf("error scanning role: %v", err)}
		}

		roles = append(roles, &role)
	}

	return roles, nil
}

func (r *repository) UpdateRole(ctx context.Context, role *models.Role) error {
	res, err := r.dbPool.Exec(ctx, r.queries.UpdateRole(), role.Name, role.Description, role.UUID)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errs.New(errs.ErrorNotFound, "role not found", nil)
	}

	return nil
}

func (r *repository) DeleteRole(ctx context.Context, uuid string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.DeleteRole(), uuid)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errs.New(errs.ErrorNotFound, "role not found", nil)
	}

	return nil
}

// Permissions

func (r *repository) CreatePermission(ctx context.Context, perm *models.Permission) error {
	_, err := r.dbPool.Exec(ctx, r.queries.CreatePermission(), perm.UUID, perm.Name, perm.Description)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	return nil
}

func (r *repository) GetAllPermissions(ctx context.Context) ([]*models.Permission, error) {
	rows, err := r.dbPool.Query(ctx, r.queries.GetAllPermissions())
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}

	var permissions []*models.Permission

	for rows.Next() {
		var perm models.Permission

		if err := rows.Scan(&perm.UUID, &perm.Name, &perm.Description, &perm.CreatedAt); err != nil {
			return nil, &errs.AppError{Code: errs.ErrorInternal, Err: fmt.Errorf("error scanning permission: %v", err)}
		}

		permissions = append(permissions, &perm)
	}

	return permissions, nil
}

func (r *repository) DeletePermission(ctx context.Context, uuid string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.DeletePermission(), uuid)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errs.New(errs.ErrorNotFound, "permission not found", nil)
	}

	return nil
}

// Role-Permission assignments

func (r *repository) AssignPermissionToRole(ctx context.Context, roleUUID, permissionUUID string) error {
	_, err := r.dbPool.Exec(ctx, r.queries.AssignPermissionToRole(), roleUUID, permissionUUID)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	return nil
}

func (r *repository) RemovePermissionFromRole(ctx context.Context, roleUUID, permissionUUID string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.RemovePermissionFromRole(), roleUUID, permissionUUID)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errs.New(errs.ErrorNotFound, "permission assignment not found", nil)
	}

	return nil
}

func (r *repository) GetRolePermissions(ctx context.Context, roleUUID string) ([]*models.Permission, error) {
	rows, err := r.dbPool.Query(ctx, r.queries.GetRolePermissions(), roleUUID)
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}

	var permissions []*models.Permission

	for rows.Next() {
		var perm models.Permission

		if err := rows.Scan(&perm.UUID, &perm.Name, &perm.Description, &perm.CreatedAt); err != nil {
			return nil, &errs.AppError{Code: errs.ErrorInternal, Err: fmt.Errorf("error scanning permission: %v", err)}
		}

		permissions = append(permissions, &perm)
	}

	return permissions, nil
}

// User-Role assignments

func (r *repository) AssignRoleToUser(ctx context.Context, userUUID, roleUUID string) error {
	_, err := r.dbPool.Exec(ctx, r.queries.AssignRoleToUser(), userUUID, roleUUID)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	return nil
}

func (r *repository) RemoveRoleFromUser(ctx context.Context, userUUID, roleUUID string) error {
	res, err := r.dbPool.Exec(ctx, r.queries.RemoveRoleFromUser(), userUUID, roleUUID)
	if err != nil {
		return errs.SQLErrorParser(err)
	}

	if res.RowsAffected() == 0 {
		return errs.New(errs.ErrorNotFound, "role assignment not found", nil)
	}

	return nil
}

func (r *repository) GetUserRoles(ctx context.Context, userUUID string) ([]*models.Role, error) {
	rows, err := r.dbPool.Query(ctx, r.queries.GetUserRoles(), userUUID)
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}

	var roles []*models.Role

	for rows.Next() {
		var role models.Role

		if err := rows.Scan(&role.UUID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, &errs.AppError{Code: errs.ErrorInternal, Err: fmt.Errorf("error scanning role: %v", err)}
		}

		roles = append(roles, &role)
	}

	return roles, nil
}

func (r *repository) GetUserPermissions(ctx context.Context, userUUID string) ([]string, error) {
	rows, err := r.dbPool.Query(ctx, r.queries.GetUserPermissions(), userUUID)
	if err != nil {
		return nil, errs.SQLErrorParser(err)
	}

	var permissions []string

	for rows.Next() {
		var perm string

		if err := rows.Scan(&perm); err != nil {
			return nil, &errs.AppError{Code: errs.ErrorInternal, Err: fmt.Errorf("error scanning permission: %v", err)}
		}

		permissions = append(permissions, perm)
	}

	return permissions, nil
}
