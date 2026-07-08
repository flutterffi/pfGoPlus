package rbac

import "slices"

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

type PermissionGroupDefinition struct {
	Key         string `json:"key"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

type RoleTemplateDefinition struct {
	Key          string   `json:"key"`
	DisplayName  string   `json:"display_name"`
	Description  string   `json:"description"`
	System       bool     `json:"system"`
	Permissions  []string `json:"permissions"`
	DefaultState string   `json:"default_state"`
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

func Groups() []PermissionGroupDefinition {
	return []PermissionGroupDefinition{
		{
			Key:         "users",
			DisplayName: "Users",
			Description: "Account directory, identity profile, and user lifecycle permissions.",
			Order:       10,
		},
		{
			Key:         "roles",
			DisplayName: "Roles",
			Description: "Role catalog, permission assignment, and authorization governance permissions.",
			Order:       20,
		},
		{
			Key:         "audit",
			DisplayName: "Audit",
			Description: "Administrative audit trail visibility and compliance review permissions.",
			Order:       30,
		},
		{
			Key:         "todos",
			DisplayName: "Todos",
			Description: "Demo business capability permissions for todo operations.",
			Order:       40,
		},
	}
}

func RoleTemplates() []RoleTemplateDefinition {
	adminPermissions := []string{
		PermissionAuditRead,
		PermissionRolesRead,
		PermissionRolesWrite,
		PermissionTodosRead,
		PermissionTodosWrite,
		PermissionUsersRead,
		PermissionUsersWrite,
	}
	memberPermissions := []string{
		PermissionTodosRead,
		PermissionTodosWrite,
	}
	slices.Sort(adminPermissions)
	slices.Sort(memberPermissions)

	return []RoleTemplateDefinition{
		{
			Key:          "admin",
			DisplayName:  "Administrator",
			Description:  "Full management template for platform administrators.",
			System:       true,
			Permissions:  adminPermissions,
			DefaultState: "active",
		},
		{
			Key:          "member",
			DisplayName:  "Member",
			Description:  "Baseline business access template for normal product users.",
			System:       true,
			Permissions:  memberPermissions,
			DefaultState: "active",
		},
	}
}

func RoleTemplateByKey(key string) (*RoleTemplateDefinition, bool) {
	for _, item := range RoleTemplates() {
		if item.Key == key {
			template := item
			return &template, true
		}
	}
	return nil, false
}
