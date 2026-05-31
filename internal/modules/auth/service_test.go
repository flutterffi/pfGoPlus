package auth

import (
	"testing"
	"time"

	"github.com/flutterffi/pfGoPlus/internal/config"
)

func newTestService() *Service {
	return NewService(config.AuthConfig{
		JWTSecret:      "test-secret",
		JWTIssuer:      "pfGoPlus-test",
		AccessTokenTTL: time.Hour,
		DemoUsername:   "admin",
		DemoPassword:   "admin123",
	})
}

func TestLoginSuccess(t *testing.T) {
	service := newTestService()

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
	service := newTestService()

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
}
