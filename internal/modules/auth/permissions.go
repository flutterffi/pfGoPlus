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

type PermissionDefinition struct {
	Key         string `json:"key"`
	Group       string `json:"group"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

func Catalog() []PermissionDefinition {
	return []PermissionDefinition{
		{
			Key:         PermissionUsersRead,
			Group:       "users",
			DisplayName: "Read Users",
			Description: "View user profiles, user lists, and account details.",
		},
		{
			Key:         PermissionUsersWrite,
			Group:       "users",
			DisplayName: "Manage Users",
			Description: "Create users and update user role, profile, password, or status.",
		},
		{
			Key:         PermissionRolesRead,
			Group:       "roles",
			DisplayName: "Read Roles",
			Description: "View role catalog, role status, and permission assignments.",
		},
		{
			Key:         PermissionRolesWrite,
			Group:       "roles",
			DisplayName: "Manage Roles",
			Description: "Create, update, disable, and delete custom roles.",
		},
		{
			Key:         PermissionAuditRead,
			Group:       "audit",
			DisplayName: "Read Audit Logs",
			Description: "View audit trails and administrative operation history.",
		},
		{
			Key:         PermissionTodosRead,
			Group:       "todos",
			DisplayName: "Read Todos",
			Description: "View todo items and todo list results.",
		},
		{
			Key:         PermissionTodosWrite,
			Group:       "todos",
			DisplayName: "Manage Todos",
			Description: "Create and update todo items through the application APIs.",
		},
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
