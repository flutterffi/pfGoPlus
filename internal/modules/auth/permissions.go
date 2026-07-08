package auth

const (
	PermissionUsersRead  = "users:read"
	PermissionUsersWrite = "users:write"
	PermissionAuditRead  = "audit:read"
	PermissionRolesRead  = "roles:read"
	PermissionRolesWrite = "roles:write"
	PermissionTodosRead  = "todos:read"
	PermissionTodosWrite = "todos:write"
)

func hasPermission(permissions []string, expected string) bool {
	for _, permission := range permissions {
		if permission == expected {
			return true
		}
	}
	return false
}
