package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/flutterffi/pfGoPlus/internal/config"
	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	secret   []byte
	issuer   string
	tokenTTL time.Duration
	username string
	password string
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
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewService(cfg config.AuthConfig) *Service {
	return &Service{
		secret:   []byte(cfg.JWTSecret),
		issuer:   cfg.JWTIssuer,
		tokenTTL: cfg.AccessTokenTTL,
		username: cfg.DemoUsername,
		password: cfg.DemoPassword,
	}
}

func (s *Service) Login(req LoginRequest) (*LoginResult, error) {
	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)

	if username == "" || password == "" {
		return nil, httpx.BadRequest("username and password are required", nil)
	}
	if username != s.username || password != s.password {
		return nil, httpx.Unauthorized("invalid username or password", nil)
	}

	expiresAt := time.Now().Add(s.tokenTTL)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   username,
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
