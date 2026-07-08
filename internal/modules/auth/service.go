package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/flutterffi/pfGoPlus/internal/config"
	"github.com/flutterffi/pfGoPlus/internal/modules/user"
	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	secret   []byte
	issuer   string
	tokenTTL time.Duration
	users    user.Repository
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type Claims struct {
	UserID      uint     `json:"user_id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func NewService(cfg config.AuthConfig, users user.Repository) *Service {
	return &Service{
		secret:   []byte(cfg.JWTSecret),
		issuer:   cfg.JWTIssuer,
		tokenTTL: cfg.AccessTokenTTL,
		users:    users,
	}
}

func (s *Service) Login(req LoginRequest) (*LoginResult, error) {
	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)

	if username == "" || password == "" {
		return nil, httpx.BadRequest("username and password are required", nil)
	}

	item, err := s.users.FindByUsername(context.Background(), username)
	if err != nil {
		return nil, httpx.Internal("load user failed", err)
	}
	if item == nil || !user.CheckPassword(item.PasswordHash, password) {
		return nil, httpx.Unauthorized("invalid username or password", nil)
	}
	if item.Status != user.StatusActive {
		return nil, httpx.Unauthorized("user is disabled", nil)
	}

	expiresAt := time.Now().Add(s.tokenTTL)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:      item.ID,
		Username:    item.Username,
		DisplayName: item.DisplayName,
		Role:        item.Role,
		Permissions: permissionsForRole(item.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   item.Username,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return nil, httpx.Internal("generate access token failed", err)
	}

	return &LoginResult{
		AccessToken: signedToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.tokenTTL.Seconds()),
	}, nil
}

func (s *Service) ParseToken(rawToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(rawToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, httpx.Unauthorized("invalid or expired access token", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, httpx.Unauthorized("invalid or expired access token", nil)
	}
	if claims.Issuer != s.issuer {
		return nil, httpx.Unauthorized("invalid token issuer", nil)
	}
	return claims, nil
}
