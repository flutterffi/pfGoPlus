package auth

import (
	"context"
	"testing"
	"time"

	"github.com/flutterffi/pfGoPlus/internal/config"
	"github.com/flutterffi/pfGoPlus/internal/modules/role"
	"github.com/flutterffi/pfGoPlus/internal/modules/user"
)

type stubUserRepo struct {
	item *user.User
}

func (s *stubUserRepo) Create(context.Context, *user.User) error { return nil }

func (s *stubUserRepo) FindByID(context.Context, uint) (*user.User, error) { return s.item, nil }

func (s *stubUserRepo) FindByUsername(_ context.Context, username string) (*user.User, error) {
	if s.item != nil && s.item.Username == username {
		return s.item, nil
	}
	return nil, nil
}

func (s *stubUserRepo) List(context.Context) ([]user.User, error) { return nil, nil }

func (s *stubUserRepo) Update(context.Context, *user.User) error { return nil }

func newTestService(t *testing.T) *Service {
	t.Helper()
	hash, err := user.HashPassword("admin123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	roleRepo := &stubRoleRepo{items: []role.Role{
		{Name: role.NameAdmin, Permissions: `["users:read","users:write","audit:read","roles:read","roles:write","todos:read","todos:write"]`},
	}}

	return NewService(config.AuthConfig{
		JWTSecret:      "test-secret",
		JWTIssuer:      "pfGoPlus-test",
		AccessTokenTTL: time.Hour,
	}, &stubUserRepo{item: &user.User{
		ID:           1,
		Username:     "admin",
		DisplayName:  "Admin",
		PasswordHash: hash,
		Role:         user.RoleAdmin,
		Status:       user.StatusActive,
	}}, role.NewService(roleRepo))
}

type stubRoleRepo struct {
	items []role.Role
}

func (s *stubRoleRepo) Create(context.Context, *role.Role) error { return nil }

func (s *stubRoleRepo) Update(_ context.Context, item *role.Role) error {
	for i := range s.items {
		if s.items[i].Name == item.Name {
			s.items[i] = *item
			return nil
		}
	}
	return nil
}

func (s *stubRoleRepo) FindByName(_ context.Context, name string) (*role.Role, error) {
	for i := range s.items {
		if s.items[i].Name == name {
			item := s.items[i]
			return &item, nil
		}
	}
	return nil, nil
}

func (s *stubRoleRepo) List(context.Context) ([]role.Role, error) { return s.items, nil }

func TestLoginSuccess(t *testing.T) {
	service := newTestService(t)

	result, err := service.Login(LoginRequest{
		Username: "admin",
		Password: "admin123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}
}

func TestParseTokenSuccess(t *testing.T) {
	service := newTestService(t)

	result, err := service.Login(LoginRequest{
		Username: "admin",
		Password: "admin123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	claims, err := service.ParseToken(result.AccessToken)
	if err != nil {
		t.Fatalf("expected valid token, got %v", err)
	}
	if claims.Username != "admin" {
		t.Fatalf("unexpected username: %s", claims.Username)
	}
	if claims.Role != user.RoleAdmin {
		t.Fatalf("unexpected role: %s", claims.Role)
	}
	if len(claims.Permissions) == 0 {
		t.Fatal("expected permissions in claims")
	}
}
