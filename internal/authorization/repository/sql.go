//revive:disable:line-length-limit
package repository

type AuthorizationQueries struct{}

// Roles
func (_ *AuthorizationQueries) CreateRole() string {
	return `INSERT INTO roles (uuid, name, description) VALUES ($1, $2, $3)`
}

func (_ *AuthorizationQueries) GetRoleByUUID() string {
	return `SELECT uuid, name, description, created_at, updated_at FROM roles WHERE uuid = $1`
}

func (_ *AuthorizationQueries) GetRoleByName() string {
	return `SELECT uuid, name, description, created_at, updated_at FROM roles WHERE name = $1`
}

func (_ *AuthorizationQueries) GetAllRoles() string {
	return `SELECT uuid, name, description, created_at, updated_at FROM roles ORDER BY name`
}

func (_ *AuthorizationQueries) UpdateRole() string {
	return `UPDATE roles SET name = $1, description = $2, updated_at = NOW() WHERE uuid = $3`
}

func (_ *AuthorizationQueries) DeleteRole() string {
	return `DELETE FROM roles WHERE uuid = $1`
}

// Permissions
func (_ *AuthorizationQueries) CreatePermission() string {
	return `INSERT INTO permissions (uuid, name, description) VALUES ($1, $2, $3)`
}

func (_ *AuthorizationQueries) GetAllPermissions() string {
	return `SELECT uuid, name, description, created_at FROM permissions ORDER BY name`
}

func (_ *AuthorizationQueries) DeletePermission() string {
	return `DELETE FROM permissions WHERE uuid = $1`
}

// Role-Permission assignments
func (_ *AuthorizationQueries) AssignPermissionToRole() string {
	return `INSERT INTO role_permissions (role_uuid, permission_uuid) VALUES ($1, $2)`
}

func (_ *AuthorizationQueries) RemovePermissionFromRole() string {
	return `DELETE FROM role_permissions WHERE role_uuid = $1 AND permission_uuid = $2`
}

func (_ *AuthorizationQueries) GetRolePermissions() string {
	return `SELECT p.uuid, p.name, p.description, p.created_at FROM permissions p JOIN role_permissions rp ON p.uuid = rp.permission_uuid WHERE rp.role_uuid = $1 ORDER BY p.name`
}

// User-Role assignments
func (_ *AuthorizationQueries) AssignRoleToUser() string {
	return `INSERT INTO user_roles (user_uuid, role_uuid) VALUES ($1, $2)`
}

func (_ *AuthorizationQueries) RemoveRoleFromUser() string {
	return `DELETE FROM user_roles WHERE user_uuid = $1 AND role_uuid = $2`
}

func (_ *AuthorizationQueries) GetUserRoles() string {
	return `SELECT r.uuid, r.name, r.description, r.created_at, r.updated_at FROM roles r JOIN user_roles ur ON r.uuid = ur.role_uuid WHERE ur.user_uuid = $1 ORDER BY r.name`
}

func (_ *AuthorizationQueries) GetUserPermissions() string {
	return `SELECT DISTINCT p.name FROM permissions p JOIN role_permissions rp ON p.uuid = rp.permission_uuid JOIN user_roles ur ON rp.role_uuid = ur.role_uuid WHERE ur.user_uuid = $1 ORDER BY p.name`
}
