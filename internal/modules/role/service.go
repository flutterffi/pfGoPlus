package role

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"github.com/flutterffi/pfGoPlus/internal/modules/rbac"
	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
)

const (
	NameAdmin  = "admin"
	NameMember = "member"

	StatusActive   = "active"
	StatusDisabled = "disabled"
)

type Service struct {
	repo        Repository
	roleCounter UsageCounter
}

type Seed struct {
	Name        string
	DisplayName string
	Permissions []string
}

type UpdateRequest struct {
	DisplayName *string  `json:"display_name"`
	Permissions []string `json:"permissions"`
	Status      *string  `json:"status"`
}

type CreateRequest struct {
	Name         string   `json:"name"`
	DisplayName  string   `json:"display_name"`
	TemplateName string   `json:"template_name"`
	Permissions  []string `json:"permissions"`
}

type UsageCounter interface {
	CountByRole(ctx context.Context, role string) (int64, error)
}

func DefaultSeeds() []Seed {
	return []Seed{
		{
			Name:        NameAdmin,
			DisplayName: "Administrator",
			Permissions: []string{"users:read", "users:write", "audit:read", "roles:read", "roles:write", "todos:read", "todos:write"},
		},
		{
			Name:        NameMember,
			DisplayName: "Member",
			Permissions: []string{"todos:read", "todos:write"},
		},
	}
}

func NewService(repo Repository, roleCounter UsageCounter) *Service {
	return &Service{repo: repo, roleCounter: roleCounter}
}

func (s *Service) EnsureDefaults(ctx context.Context) error {
	for _, seed := range DefaultSeeds() {
		existing, err := s.repo.FindByName(ctx, seed.Name)
		if err != nil {
			return httpx.Internal("load role failed", err)
		}
		if existing != nil {
			continue
		}
		encoded, err := json.Marshal(seed.Permissions)
		if err != nil {
			return httpx.Internal("encode role permissions failed", err)
		}
		if err := s.repo.Create(ctx, &Role{
			Name:        seed.Name,
			DisplayName: seed.DisplayName,
			Permissions: string(encoded),
			Status:      StatusActive,
		}); err != nil {
			return httpx.Internal("create default role failed", err)
		}
	}
	return nil
}

func (s *Service) ResolvePermissions(ctx context.Context, name string) ([]string, error) {
	item, err := s.repo.FindByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return nil, httpx.Internal("load role failed", err)
	}
	if item == nil {
		return nil, httpx.BadRequest("role not found", nil)
	}
	var permissions []string
	if err := json.Unmarshal([]byte(item.Permissions), &permissions); err != nil {
		return nil, httpx.Internal("decode role permissions failed", err)
	}
	if item.Status != StatusActive {
		return nil, httpx.Forbidden("role is disabled", nil)
	}
	return permissions, nil
}

func (s *Service) RoleExists(ctx context.Context, name string) (bool, error) {
	item, err := s.repo.FindByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return false, httpx.Internal("load role failed", err)
	}
	return item != nil, nil
}

func (s *Service) List(ctx context.Context) ([]Role, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, httpx.Internal("list roles failed", err)
	}
	return items, nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Role, error) {
	name := strings.ToLower(strings.TrimSpace(req.Name))
	displayName := strings.TrimSpace(req.DisplayName)
	if name == "" {
		return nil, httpx.BadRequest("name is required", nil)
	}
	if displayName == "" {
		displayName = name
	}

	existing, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, httpx.Internal("load role failed", err)
	}
	if existing != nil {
		return nil, httpx.BadRequest("role already exists", nil)
	}

	var (
		permissions []string
		validateErr error
	)
	if len(req.Permissions) == 0 {
		permissions, validateErr = resolveTemplatePermissions(req.TemplateName)
	} else {
		permissions, validateErr = normalizePermissions(req.Permissions)
	}
	if validateErr != nil {
		return nil, validateErr
	}
	encoded, err := json.Marshal(permissions)
	if err != nil {
		return nil, httpx.Internal("encode role permissions failed", err)
	}

	item := &Role{
		Name:        name,
		DisplayName: displayName,
		Permissions: string(encoded),
		Status:      StatusActive,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, httpx.Internal("create role failed", err)
	}
	return item, nil
}

func (s *Service) Update(ctx context.Context, name string, req UpdateRequest) (*Role, error) {
	item, err := s.repo.FindByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return nil, httpx.Internal("load role failed", err)
	}
	if item == nil {
		return nil, httpx.NotFound("role not found", nil)
	}

	if req.DisplayName != nil {
		displayName := strings.TrimSpace(*req.DisplayName)
		if displayName == "" {
			return nil, httpx.BadRequest("display_name cannot be empty", nil)
		}
		item.DisplayName = displayName
	}

	if req.Permissions != nil {
		permissions, validateErr := normalizePermissions(req.Permissions)
		if validateErr != nil {
			return nil, validateErr
		}
		if item.Name == NameAdmin {
			for _, permission := range []string{"users:read", "users:write", "audit:read", "roles:read", "roles:write"} {
				if !slices.Contains(permissions, permission) {
					return nil, httpx.BadRequest("admin role must keep core management permissions", nil)
				}
			}
		}
		encoded, err := json.Marshal(permissions)
		if err != nil {
			return nil, httpx.Internal("encode role permissions failed", err)
		}
		item.Permissions = string(encoded)
	}

	if req.Status != nil {
		status := normalizeStatus(*req.Status)
		if !isValidStatus(status) {
			return nil, httpx.BadRequest("status must be active or disabled", nil)
		}
		if item.Name == NameAdmin && status != StatusActive {
			return nil, httpx.BadRequest("admin role cannot be disabled", nil)
		}
		if item.Status == StatusActive && status == StatusDisabled && s.roleCounter != nil {
			count, err := s.roleCounter.CountByRole(ctx, item.Name)
			if err != nil {
				return nil, httpx.Internal("count role usage failed", err)
			}
			if count > 0 {
				return nil, httpx.BadRequest("role is assigned to existing users", nil)
			}
		}
		item.Status = status
	}

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, httpx.Internal("update role failed", err)
	}
	return item, nil
}

func (s *Service) Delete(ctx context.Context, name string) error {
	item, err := s.repo.FindByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return httpx.Internal("load role failed", err)
	}
	if item == nil {
		return httpx.NotFound("role not found", nil)
	}
	if item.Name == NameAdmin || item.Name == NameMember {
		return httpx.BadRequest("system role cannot be deleted", nil)
	}
	if item.Status != StatusDisabled {
		return httpx.BadRequest("role must be disabled before deletion", nil)
	}
	if s.roleCounter != nil {
		count, err := s.roleCounter.CountByRole(ctx, item.Name)
		if err != nil {
			return httpx.Internal("count role usage failed", err)
		}
		if count > 0 {
			return httpx.BadRequest("role is assigned to existing users", nil)
		}
	}
	if err := s.repo.Delete(ctx, item.Name); err != nil {
		return httpx.Internal("delete role failed", err)
	}
	return nil
}

func resolveTemplatePermissions(templateName string) ([]string, error) {
	templateName = strings.ToLower(strings.TrimSpace(templateName))
	if templateName == "" {
		return nil, httpx.BadRequest("permissions cannot be empty", nil)
	}
	template, ok := rbac.RoleTemplateByKey(templateName)
	if !ok {
		return nil, httpx.BadRequest("role template not found", nil)
	}
	permissions := make([]string, len(template.Permissions))
	copy(permissions, template.Permissions)
	return permissions, nil
}

func normalizePermissions(values []string) ([]string, error) {
	seen := make(map[string]struct{}, len(values))
	permissions := make([]string, 0, len(values))
	for _, value := range values {
		permission := strings.TrimSpace(value)
		if permission == "" {
			continue
		}
		if _, ok := seen[permission]; ok {
			continue
		}
		seen[permission] = struct{}{}
		permissions = append(permissions, permission)
	}
	if len(permissions) == 0 {
		return nil, httpx.BadRequest("permissions cannot be empty", nil)
	}
	slices.Sort(permissions)
	return permissions, nil
}

func normalizeStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", StatusActive:
		return StatusActive
	case StatusDisabled:
		return StatusDisabled
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func isValidStatus(status string) bool {
	return status == StatusActive || status == StatusDisabled
}
