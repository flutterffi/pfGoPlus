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

func (s *stubRepository) FindByID(_ context.Context, id uint) (*User, error) {
	for i := range s.items {
		if s.items[i].ID == id {
			item := s.items[i]
			return &item, nil
		}
	}
	return nil, nil
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

func (s *stubRepository) Update(_ context.Context, item *User) error {
	for i := range s.items {
		if s.items[i].ID == item.ID {
			item.UpdatedAt = time.Date(2026, 1, 2, 4, 5, 6, 0, time.UTC)
			s.items[i] = *item
			return nil
		}
	}
	return nil
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

func TestUpdateUserDisableSuccess(t *testing.T) {
	repo := &stubRepository{items: []User{{
		ID:           1,
		Username:     "alice",
		DisplayName:  "Alice",
		PasswordHash: "hashed",
		Role:         RoleMember,
		Status:       StatusActive,
	}}}
	service, err := NewService(config.AuthConfig{}, repo)
	if err != nil {
		t.Fatalf("new user service: %v", err)
	}

	status := StatusDisabled
	item, err := service.Update(context.Background(), 1, UpdateRequest{Status: &status}, 99)
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if item.Status != StatusDisabled {
		t.Fatalf("expected disabled status, got %s", item.Status)
	}
}

func TestUpdateUserCannotDisableSelf(t *testing.T) {
	repo := &stubRepository{items: []User{{
		ID:           1,
		Username:     "admin",
		DisplayName:  "Admin",
		PasswordHash: "hashed",
		Role:         RoleAdmin,
		Status:       StatusActive,
	}}}
	service, err := NewService(config.AuthConfig{}, repo)
	if err != nil {
		t.Fatalf("new user service: %v", err)
	}

	status := StatusDisabled
	_, err = service.Update(context.Background(), 1, UpdateRequest{Status: &status}, 1)
	if err == nil {
		t.Fatal("expected self-disable error")
	}
}

func TestUpdateUserCannotRemoveOwnAdminRole(t *testing.T) {
	repo := &stubRepository{items: []User{{
		ID:           1,
		Username:     "admin",
		DisplayName:  "Admin",
		PasswordHash: "hashed",
		Role:         RoleAdmin,
		Status:       StatusActive,
	}}}
	service, err := NewService(config.AuthConfig{}, repo)
	if err != nil {
		t.Fatalf("new user service: %v", err)
	}

	role := RoleMember
	_, err = service.Update(context.Background(), 1, UpdateRequest{Role: &role}, 1)
	if err == nil {
		t.Fatal("expected self-demotion error")
	}
}
