package role

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
)

const (
	NameAdmin  = "admin"
	NameMember = "member"
)

type Service struct {
	repo Repository
}

type Seed struct {
	Name        string
	DisplayName string
	Permissions []string
}

type UpdateRequest struct {
	DisplayName *string  `json:"display_name"`
	Permissions []string `json:"permissions"`
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

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
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

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, httpx.Internal("update role failed", err)
	}
	return item, nil
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
