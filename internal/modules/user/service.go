package user

import (
	"context"
	"strings"

	"github.com/flutterffi/pfGoPlus/internal/config"
	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
)

type Service struct {
	repo Repository
}

type CreateRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
	Role        string `json:"role"`
}

func NewService(cfg config.AuthConfig, repo Repository) (*Service, error) {
	service := &Service{repo: repo}
	if err := service.EnsureBootstrapAdmin(context.Background(), cfg); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *Service) EnsureBootstrapAdmin(ctx context.Context, cfg config.AuthConfig) error {
	username := strings.TrimSpace(cfg.DemoUsername)
	password := strings.TrimSpace(cfg.DemoPassword)
	if username == "" || password == "" {
		return nil
	}

	existing, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return httpx.Internal("load bootstrap admin failed", err)
	}
	if existing != nil {
		return nil
	}

	hash, err := HashPassword(password)
	if err != nil {
		return httpx.Internal("hash bootstrap admin password failed", err)
	}

	item := &User{
		Username:     username,
		DisplayName:  username,
		PasswordHash: hash,
		Role:         RoleAdmin,
		Status:       StatusActive,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return httpx.Internal("create bootstrap admin failed", err)
	}
	return nil
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*User, error) {
	username := strings.TrimSpace(req.Username)
	displayName := strings.TrimSpace(req.DisplayName)
	password := strings.TrimSpace(req.Password)
	role := normalizeRole(req.Role)

	if username == "" || password == "" {
		return nil, httpx.BadRequest("username and password are required", nil)
	}
	if displayName == "" {
		displayName = username
	}
	if !isValidRole(role) {
		return nil, httpx.BadRequest("role must be admin or member", nil)
	}

	existing, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, httpx.Internal("check existing user failed", err)
	}
	if existing != nil {
		return nil, httpx.BadRequest("username already exists", nil)
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, httpx.Internal("hash user password failed", err)
	}

	item := &User{
		Username:     username,
		DisplayName:  displayName,
		PasswordHash: hash,
		Role:         role,
		Status:       StatusActive,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, httpx.Internal("create user failed", err)
	}
	return item, nil
}

func (s *Service) List(ctx context.Context) ([]User, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, httpx.Internal("list users failed", err)
	}
	return items, nil
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "", RoleMember:
		return RoleMember
	case RoleAdmin:
		return RoleAdmin
	default:
		return strings.ToLower(strings.TrimSpace(role))
	}
}

func isValidRole(role string) bool {
	return role == RoleAdmin || role == RoleMember
}
