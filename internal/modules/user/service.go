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

type UpdateRequest struct {
	DisplayName *string `json:"display_name"`
	Password    *string `json:"password"`
	Role        *string `json:"role"`
	Status      *string `json:"status"`
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

func (s *Service) Update(ctx context.Context, id uint, req UpdateRequest, actorID uint) (*User, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, httpx.Internal("load user failed", err)
	}
	if item == nil {
		return nil, httpx.NotFound("user not found", nil)
	}

	if req.DisplayName != nil {
		displayName := strings.TrimSpace(*req.DisplayName)
		if displayName == "" {
			return nil, httpx.BadRequest("display_name cannot be empty", nil)
		}
		item.DisplayName = displayName
	}

	if req.Password != nil {
		password := strings.TrimSpace(*req.Password)
		if password == "" {
			return nil, httpx.BadRequest("password cannot be empty", nil)
		}
		hash, err := HashPassword(password)
		if err != nil {
			return nil, httpx.Internal("hash user password failed", err)
		}
		item.PasswordHash = hash
	}

	if req.Role != nil {
		role := normalizeRole(*req.Role)
		if !isValidRole(role) {
			return nil, httpx.BadRequest("role must be admin or member", nil)
		}
		if actorID == item.ID && item.Role == RoleAdmin && role != RoleAdmin {
			return nil, httpx.BadRequest("admin cannot remove own admin role", nil)
		}
		item.Role = role
	}

	if req.Status != nil {
		status := normalizeStatus(*req.Status)
		if !isValidStatus(status) {
			return nil, httpx.BadRequest("status must be active or disabled", nil)
		}
		if actorID == item.ID && item.Status == StatusActive && status != StatusActive {
			return nil, httpx.BadRequest("user cannot disable self", nil)
		}
		item.Status = status
	}

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, httpx.Internal("update user failed", err)
	}
	return item, nil
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

func normalizeStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", StatusActive:
		return StatusActive
	case StatusDisabled:
		return StatusDisabled
	default:
		return strings.ToLower(strings.TrimSpace(status))
	}
}

func isValidStatus(status string) bool {
	return status == StatusActive || status == StatusDisabled
}
