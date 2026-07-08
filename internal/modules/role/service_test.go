package role

import (
	"context"
	"testing"
)

type stubRepository struct {
	items []Role
}

func (s *stubRepository) Create(_ context.Context, item *Role) error {
	item.ID = uint(len(s.items) + 1)
	s.items = append(s.items, *item)
	return nil
}

func (s *stubRepository) FindByName(_ context.Context, name string) (*Role, error) {
	for i := range s.items {
		if s.items[i].Name == name {
			item := s.items[i]
			return &item, nil
		}
	}
	return nil, nil
}

func (s *stubRepository) List(_ context.Context) ([]Role, error) {
	items := make([]Role, len(s.items))
	copy(items, s.items)
	return items, nil
}

func (s *stubRepository) Update(_ context.Context, item *Role) error {
	for i := range s.items {
		if s.items[i].Name == item.Name {
			s.items[i] = *item
			return nil
		}
	}
	return nil
}

func TestEnsureDefaultsSeedsRoles(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)

	if err := service.EnsureDefaults(context.Background()); err != nil {
		t.Fatalf("ensure defaults: %v", err)
	}
	if len(repo.items) != 2 {
		t.Fatalf("expected 2 seeded roles, got %d", len(repo.items))
	}
}

func TestResolvePermissions(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)
	if err := service.EnsureDefaults(context.Background()); err != nil {
		t.Fatalf("ensure defaults: %v", err)
	}

	permissions, err := service.ResolvePermissions(context.Background(), NameAdmin)
	if err != nil {
		t.Fatalf("resolve permissions: %v", err)
	}
	if len(permissions) == 0 {
		t.Fatal("expected permissions")
	}
}

func TestUpdateRolePermissions(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)
	if err := service.EnsureDefaults(context.Background()); err != nil {
		t.Fatalf("ensure defaults: %v", err)
	}

	displayName := "Platform Admin"
	item, err := service.Update(context.Background(), NameAdmin, UpdateRequest{
		DisplayName: &displayName,
		Permissions: []string{"users:read", "users:write", "audit:read", "roles:read", "roles:write", "todos:read"},
	})
	if err != nil {
		t.Fatalf("update role: %v", err)
	}
	if item.DisplayName != "Platform Admin" {
		t.Fatalf("unexpected display name: %s", item.DisplayName)
	}
}

func TestUpdateAdminRoleKeepsCorePermissions(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)
	if err := service.EnsureDefaults(context.Background()); err != nil {
		t.Fatalf("ensure defaults: %v", err)
	}

	_, err := service.Update(context.Background(), NameAdmin, UpdateRequest{
		Permissions: []string{"todos:read"},
	})
	if err == nil {
		t.Fatal("expected admin core permissions validation error")
	}
}
