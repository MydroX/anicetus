//revive:disable:max-public-structs
package dto

type CreateRoleRequest struct {
	Name        string `json:"name"        validate:"required,min=2,max=50"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"        validate:"required,min=2,max=50"`
	Description string `json:"description"`
}

type CreatePermissionRequest struct {
	Name        string `json:"name"        validate:"required,min=2,max=100"`
	Description string `json:"description"`
}

type AssignPermissionToRoleRequest struct {
	PermissionUUID string `json:"permission_uuid" validate:"required"`
}

type AssignRoleToUserRequest struct {
	RoleUUID string `json:"role_uuid" validate:"required"`
}

type RoleResponse struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PermissionResponse struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UserRolesResponse struct {
	Roles []*RoleResponse `json:"roles"`
}

type UserPermissionsResponse struct {
	Permissions []string `json:"permissions"`
}
