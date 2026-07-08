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
