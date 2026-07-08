package role

import (
	"context"
	"testing"
)

type stubRepository struct {
	items []Role
}

type stubUsageCounter struct {
	counts map[string]int64
}

func (s *stubRepository) Create(_ context.Context, item *Role) error {
	item.ID = uint(len(s.items) + 1)
	s.items = append(s.items, *item)
	return nil
}

func (s *stubRepository) Delete(_ context.Context, name string) error {
	for i := range s.items {
		if s.items[i].Name == name {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return nil
		}
	}
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

func (s *stubUsageCounter) CountByRole(_ context.Context, role string) (int64, error) {
	return s.counts[role], nil
}

func TestEnsureDefaultsSeedsRoles(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo, &stubUsageCounter{counts: map[string]int64{}})

	if err := service.EnsureDefaults(context.Background()); err != nil {
		t.Fatalf("ensure defaults: %v", err)
	}
	if len(repo.items) != 2 {
		t.Fatalf("expected 2 seeded roles, got %d", len(repo.items))
	}
}

func TestResolvePermissions(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo, &stubUsageCounter{counts: map[string]int64{}})
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
	service := NewService(repo, &stubUsageCounter{counts: map[string]int64{}})
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
	service := NewService(repo, &stubUsageCounter{counts: map[string]int64{}})
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

func TestCreateRoleSuccess(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo, &stubUsageCounter{counts: map[string]int64{}})

	item, err := service.Create(context.Background(), CreateRequest{
		Name:        "auditor",
		DisplayName: "Auditor",
		Permissions: []string{"audit:read"},
	})
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	if item.Name != "auditor" {
		t.Fatalf("unexpected role name: %s", item.Name)
	}
}

func TestCreateRoleFromTemplateSuccess(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo, &stubUsageCounter{counts: map[string]int64{}})

	item, err := service.Create(context.Background(), CreateRequest{
		Name:         "operator",
		DisplayName:  "Operator",
		TemplateName: "member",
	})
	if err != nil {
		t.Fatalf("create role from template: %v", err)
	}

	permissions, err := service.ResolvePermissions(context.Background(), item.Name)
	if err != nil {
		t.Fatalf("resolve permissions: %v", err)
	}
	if len(permissions) != 2 {
		t.Fatalf("expected template permissions, got %d", len(permissions))
	}
}

func TestDisableRoleBlockedWhenAssigned(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo, &stubUsageCounter{counts: map[string]int64{NameMember: 2}})
	if err := service.EnsureDefaults(context.Background()); err != nil {
		t.Fatalf("ensure defaults: %v", err)
	}

	status := StatusDisabled
	_, err := service.Update(context.Background(), NameMember, UpdateRequest{Status: &status})
	if err == nil {
		t.Fatal("expected assigned role disable error")
	}
}

func TestDeleteRoleSuccess(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo, &stubUsageCounter{counts: map[string]int64{}})
	if err := service.EnsureDefaults(context.Background()); err != nil {
		t.Fatalf("ensure defaults: %v", err)
	}
	if _, err := service.Create(context.Background(), CreateRequest{
		Name:        "auditor",
		DisplayName: "Auditor",
		Permissions: []string{"audit:read"},
	}); err != nil {
		t.Fatalf("create role: %v", err)
	}

	status := StatusDisabled
	if _, err := service.Update(context.Background(), "auditor", UpdateRequest{Status: &status}); err != nil {
		t.Fatalf("disable role: %v", err)
	}
	if err := service.Delete(context.Background(), "auditor"); err != nil {
		t.Fatalf("delete role: %v", err)
	}

	exists, err := service.RoleExists(context.Background(), "auditor")
	if err != nil {
		t.Fatalf("check role exists: %v", err)
	}
	if exists {
		t.Fatal("expected deleted role to be removed")
	}
}

func TestDeleteRoleRequiresDisabledStatus(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo, &stubUsageCounter{counts: map[string]int64{}})
	if err := service.EnsureDefaults(context.Background()); err != nil {
		t.Fatalf("ensure defaults: %v", err)
	}
	if _, err := service.Create(context.Background(), CreateRequest{
		Name:        "auditor",
		DisplayName: "Auditor",
		Permissions: []string{"audit:read"},
	}); err != nil {
		t.Fatalf("create role: %v", err)
	}

	err := service.Delete(context.Background(), "auditor")
	if err == nil {
		t.Fatal("expected active role delete error")
	}
}

func TestDeleteSystemRoleBlocked(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo, &stubUsageCounter{counts: map[string]int64{}})
	if err := service.EnsureDefaults(context.Background()); err != nil {
		t.Fatalf("ensure defaults: %v", err)
	}

	err := service.Delete(context.Background(), NameAdmin)
	if err == nil {
		t.Fatal("expected system role delete error")
	}
}
