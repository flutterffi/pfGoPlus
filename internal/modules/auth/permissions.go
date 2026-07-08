package auth

import "github.com/flutterffi/pfGoPlus/internal/modules/rbac"

const (
	PermissionUsersRead  = rbac.PermissionUsersRead
	PermissionUsersWrite = rbac.PermissionUsersWrite
	PermissionAuditRead  = rbac.PermissionAuditRead
	PermissionRolesRead  = rbac.PermissionRolesRead
	PermissionRolesWrite = rbac.PermissionRolesWrite
	PermissionTodosRead  = rbac.PermissionTodosRead
	PermissionTodosWrite = rbac.PermissionTodosWrite
)

type PermissionDefinition = rbac.PermissionDefinition
type PermissionGroupDefinition = rbac.PermissionGroupDefinition
type RoleTemplateDefinition = rbac.RoleTemplateDefinition

func Catalog() []PermissionDefinition {
	return rbac.Catalog()
}

func Groups() []PermissionGroupDefinition {
	return rbac.Groups()
}

func RoleTemplates() []RoleTemplateDefinition {
	return rbac.RoleTemplates()
}

func hasPermission(permissions []string, expected string) bool {
	for _, permission := range permissions {
		if permission == expected {
			return true
		}
	}
	return false
}
