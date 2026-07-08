package auth

import "strings"

const (
	PermissionUsersRead  = "users:read"
	PermissionUsersWrite = "users:write"
	PermissionAuditRead  = "audit:read"
	PermissionTodosRead  = "todos:read"
	PermissionTodosWrite = "todos:write"
)

func permissionsForRole(role string) []string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin":
		return []string{
			PermissionUsersRead,
			PermissionUsersWrite,
			PermissionAuditRead,
			PermissionTodosRead,
			PermissionTodosWrite,
		}
	case "member":
		return []string{
			PermissionTodosRead,
			PermissionTodosWrite,
		}
	default:
		return nil
	}
}

func hasPermission(permissions []string, expected string) bool {
	for _, permission := range permissions {
		if permission == expected {
			return true
		}
	}
	return false
}
