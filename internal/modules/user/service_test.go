package user

import (
	"context"
	"testing"
	"time"

	"github.com/flutterffi/pfGoPlus/internal/config"
)

type stubRepository struct {
	items []User
}

func (s *stubRepository) Create(_ context.Context, item *User) error {
	item.ID = uint(len(s.items) + 1)
	item.CreatedAt = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	item.UpdatedAt = item.CreatedAt
	s.items = append(s.items, *item)
	return nil
}

func (s *stubRepository) FindByUsername(_ context.Context, username string) (*User, error) {
	for i := range s.items {
		if s.items[i].Username == username {
			item := s.items[i]
			return &item, nil
		}
	}
	return nil, nil
}

func (s *stubRepository) List(_ context.Context) ([]User, error) {
	items := make([]User, len(s.items))
	copy(items, s.items)
	return items, nil
}

func TestNewServiceSeedsBootstrapAdmin(t *testing.T) {
	repo := &stubRepository{}

	service, err := NewService(config.AuthConfig{
		DemoUsername: "admin",
		DemoPassword: "admin123",
	}, repo)
	if err != nil {
		t.Fatalf("new user service: %v", err)
	}
	if service == nil {
		t.Fatal("expected service")
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected 1 seeded user, got %d", len(repo.items))
	}
	if repo.items[0].Role != RoleAdmin {
		t.Fatalf("expected admin role, got %s", repo.items[0].Role)
	}
	if !CheckPassword(repo.items[0].PasswordHash, "admin123") {
		t.Fatal("expected bootstrap password to be hashed and verifiable")
	}
}

func TestCreateUserSuccess(t *testing.T) {
	repo := &stubRepository{}
	service, err := NewService(config.AuthConfig{}, repo)
	if err != nil {
		t.Fatalf("new user service: %v", err)
	}

	item, err := service.Create(context.Background(), CreateRequest{
		Username:    "alice",
		DisplayName: "Alice",
		Password:    "secret123",
		Role:        RoleMember,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if item.Role != RoleMember {
		t.Fatalf("expected member role, got %s", item.Role)
	}
	if item.PasswordHash == "secret123" {
		t.Fatal("expected password to be hashed")
	}
}
