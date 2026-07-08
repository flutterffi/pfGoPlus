package auth

import (
	"context"
	"testing"
	"time"

	"github.com/flutterffi/pfGoPlus/internal/config"
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
	}})
}

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
