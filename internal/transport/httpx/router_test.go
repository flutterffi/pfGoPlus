package httpx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flutterffi/pfGoPlus/internal/bff"
	"github.com/flutterffi/pfGoPlus/internal/config"
	"github.com/flutterffi/pfGoPlus/internal/modules/auth"
	"github.com/flutterffi/pfGoPlus/internal/modules/todo"
	"github.com/flutterffi/pfGoPlus/internal/platform/telemetry"
	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"go.uber.org/zap"
)

func TestHealthEndpoint(t *testing.T) {
	router := newTestRouter(t)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Trace-ID") == "" {
		t.Fatal("expected trace id header")
	}
	if recorder.Header().Get("X-Otel-Trace-ID") == "" {
		t.Fatal("expected otel trace id header")
	}
}

func TestMetaEndpoint(t *testing.T) {
	router := newTestRouter(t)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/meta", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response struct {
		Data struct {
			App         string `json:"app"`
			Version     string `json:"version"`
			TodoBackend string `json:"todo_backend"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal meta response: %v", err)
	}
	if response.Data.App != "pfGoPlus-test" {
		t.Fatalf("unexpected app: %s", response.Data.App)
	}
	if response.Data.TodoBackend != "local" {
		t.Fatalf("unexpected todo_backend: %s", response.Data.TodoBackend)
	}
}

func TestMetricsEndpointUnavailableInStdoutMode(t *testing.T) {
	router := newTestRouter(t)

	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

func TestProtectedTodoEndpointRequiresAuth(t *testing.T) {
	router := newTestRouter(t)

	body, _ := json.Marshal(map[string]string{
		"title":       "Ship API",
		"description": "add basic endpoint",
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/todos", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}
}

func TestCreateTodoEndpoint(t *testing.T) {
	router := newTestRouter(t)
	token := loginToken(t, router)

	body, _ := json.Marshal(map[string]string{
		"title":       "Ship API",
		"description": "add basic endpoint",
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/todos", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	cfg := config.Config{
		App: config.AppConfig{
			Name: "pfGoPlus-test",
			Env:  "test",
		},
		GRPC: config.GRPCConfig{
			ClientTarget: "127.0.0.1:9090",
		},
		Auth: config.AuthConfig{
			JWTSecret:      "test-secret",
			JWTIssuer:      "pfGoPlus-test",
			AccessTokenTTL: time.Hour,
			DemoUsername:   "admin",
			DemoPassword:   "admin123",
		},
		Observability: config.ObservabilityConfig{
			Exporter:       "stdout",
			MetricsPath:    "/metrics",
			ServiceVersion: "test-version",
		},
		TodoBackend: config.TodoBackendConfig{
			Mode: "local",
		},
	}
	authService := auth.NewService(cfg.Auth)
	authHandler := auth.NewHandler(authService)
	todoHandler := todo.NewHandler(todo.NewHTTPAdapter(todo.NewService(&fakeTodoRepo{})), auth.RequireAuth(authService))
	telemetryProvider := telemetry.NewNoop("pfGoPlus-test")
	edge := bff.New(cfg, authHandler, todoHandler, telemetryProvider)
	return httpx.NewRouter(zap.NewNop(), telemetryProvider, edge)
}

func loginToken(t *testing.T, router http.Handler) string {
	t.Helper()

	body, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": "admin123",
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d", recorder.Code)
	}

	var response struct {
		Data struct {
			Token struct {
				AccessToken string `json:"access_token"`
			} `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}
	if response.Data.Token.AccessToken == "" {
		t.Fatal("expected access token in login response")
	}
	return response.Data.Token.AccessToken
}

type fakeTodoRepo struct{}

func (f *fakeTodoRepo) Create(_ context.Context, item *todo.Todo) error {
	item.ID = 1
	return nil
}

func (f *fakeTodoRepo) List(_ context.Context) ([]todo.Todo, error) {
	return []todo.Todo{}, nil
}
