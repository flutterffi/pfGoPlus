package role

import (
	"context"
	"encoding/json"
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

func DefaultSeeds() []Seed {
	return []Seed{
		{
			Name:        NameAdmin,
			DisplayName: "Administrator",
			Permissions: []string{"users:read", "users:write", "audit:read", "roles:read", "todos:read", "todos:write"},
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
