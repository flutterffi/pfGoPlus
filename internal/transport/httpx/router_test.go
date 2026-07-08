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
	"github.com/flutterffi/pfGoPlus/internal/modules/audit"
	"github.com/flutterffi/pfGoPlus/internal/modules/auth"
	"github.com/flutterffi/pfGoPlus/internal/modules/role"
	"github.com/flutterffi/pfGoPlus/internal/modules/todo"
	"github.com/flutterffi/pfGoPlus/internal/modules/user"
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

func TestCurrentUserEndpoint(t *testing.T) {
	router := newTestRouter(t)
	token := loginToken(t, router)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response struct {
		Data struct {
			User struct {
				Permissions []string `json:"permissions"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal me response: %v", err)
	}
	if len(response.Data.User.Permissions) == 0 {
		t.Fatal("expected permissions in current user response")
	}
}

func TestAdminCanCreateUser(t *testing.T) {
	router := newTestRouter(t)
	token := loginToken(t, router)

	body, _ := json.Marshal(map[string]string{
		"username":     "alice",
		"display_name": "Alice",
		"password":     "secret123",
		"role":         "member",
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}
}

func TestAdminCanDisableUser(t *testing.T) {
	router := newTestRouter(t)
	token := loginToken(t, router)

	body, _ := json.Marshal(map[string]string{
		"username":     "alice",
		"display_name": "Alice",
		"password":     "secret123",
		"role":         "member",
	})
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	createRequest.Header.Set("Content-Type", "application/json")
	createRequest.Header.Set("Authorization", "Bearer "+token)
	createRecorder := httptest.NewRecorder()
	router.ServeHTTP(createRecorder, createRequest)

	patchBody, _ := json.Marshal(map[string]string{
		"status": "disabled",
	})
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/users/2", bytes.NewReader(patchBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

func TestMemberCannotListUsers(t *testing.T) {
	router := newMemberRouter(t)
	token := loginTokenWithCredentials(t, router, "member", "member123")

	request := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", recorder.Code)
	}
}

func TestAdminCanListAuditLogs(t *testing.T) {
	router := newTestRouter(t)
	token := loginToken(t, router)

	body, _ := json.Marshal(map[string]string{
		"username":     "alice",
		"display_name": "Alice",
		"password":     "secret123",
		"role":         "member",
	})
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	createRequest.Header.Set("Content-Type", "application/json")
	createRequest.Header.Set("Authorization", "Bearer "+token)
	createRecorder := httptest.NewRecorder()
	router.ServeHTTP(createRecorder, createRequest)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/audit/logs?action=user.create&resource=user&limit=10&offset=0", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response struct {
		Data struct {
			Total int `json:"total"`
			Items []struct {
				Action string `json:"action"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal audit response: %v", err)
	}
	if len(response.Data.Items) == 0 {
		t.Fatal("expected at least one audit log")
	}
	if response.Data.Total == 0 {
		t.Fatal("expected total count")
	}
	for _, item := range response.Data.Items {
		if item.Action != "user.create" {
			t.Fatalf("expected filtered action user.create, got %s", item.Action)
		}
	}
}

func TestAdminCanListRoles(t *testing.T) {
	router := newTestRouter(t)
	token := loginToken(t, router)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/roles", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response struct {
		Data struct {
			Items []struct {
				Name string `json:"name"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal roles response: %v", err)
	}
	if len(response.Data.Items) < 2 {
		t.Fatalf("expected at least 2 roles, got %d", len(response.Data.Items))
	}
}

func TestAdminCanUpdateRole(t *testing.T) {
	router := newTestRouter(t)
	token := loginToken(t, router)

	body, _ := json.Marshal(map[string]any{
		"display_name": "Platform Admin",
		"permissions":  []string{"users:read", "users:write", "audit:read", "roles:read", "roles:write", "todos:read"},
	})
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/roles/admin", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

func TestAdminCanCreateRole(t *testing.T) {
	router := newTestRouter(t)
	token := loginToken(t, router)

	body, _ := json.Marshal(map[string]any{
		"name":         "auditor",
		"display_name": "Auditor",
		"permissions":  []string{"audit:read"},
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader(body))
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
	userRepo := &fakeUserRepo{}
	auditRepo := &fakeAuditRepo{}
	roleRepo := newFakeRoleRepo()
	roleService := role.NewService(roleRepo, userRepo)
	if err := roleService.EnsureDefaults(context.Background()); err != nil {
		t.Fatalf("ensure default roles: %v", err)
	}
	userService, err := user.NewService(cfg.Auth, userRepo, roleService)
	if err != nil {
		t.Fatalf("new user service: %v", err)
	}
	auditService := audit.NewService(auditRepo)
	authService := auth.NewService(cfg.Auth, userRepo, roleService)
	authHandler := auth.NewHandler(authService)
	roleHandler := role.NewHandler(
		roleService,
		auth.RequirePermission(authService, auth.PermissionRolesRead),
		auth.RequirePermission(authService, auth.PermissionRolesWrite),
	)
	auditHandler := audit.NewHandler(auditService, auth.RequirePermission(authService, auth.PermissionAuditRead))
	userHandler := user.NewHandler(
		userService,
		auditService,
		auth.RequireAuth(authService),
		auth.RequirePermission(authService, auth.PermissionUsersRead),
		auth.RequirePermission(authService, auth.PermissionUsersWrite),
	)
	todoHandler := todo.NewHandler(todo.NewHTTPAdapter(todo.NewService(&fakeTodoRepo{})), auth.RequireAuth(authService))
	telemetryProvider := telemetry.NewNoop("pfGoPlus-test")
	edge := bff.New(cfg, authHandler, roleHandler, auditHandler, userHandler, todoHandler, telemetryProvider)
	return httpx.NewRouter(zap.NewNop(), telemetryProvider, edge)
}

func loginToken(t *testing.T, router http.Handler) string {
	t.Helper()
	return loginTokenWithCredentials(t, router, "admin", "admin123")
}

func loginTokenWithCredentials(t *testing.T, router http.Handler, username, password string) string {
	t.Helper()

	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
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

func newMemberRouter(t *testing.T) http.Handler {
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
	userRepo := &fakeUserRepo{}
	auditRepo := &fakeAuditRepo{}
	roleRepo := newFakeRoleRepo()
	roleService := role.NewService(roleRepo, userRepo)
	if err := roleService.EnsureDefaults(context.Background()); err != nil {
		t.Fatalf("ensure default roles: %v", err)
	}
	userService, err := user.NewService(cfg.Auth, userRepo, roleService)
	if err != nil {
		t.Fatalf("new user service: %v", err)
	}
	if _, err := userService.Create(context.Background(), user.CreateRequest{
		Username:    "member",
		DisplayName: "Member",
		Password:    "member123",
		Role:        user.RoleMember,
	}); err != nil {
		t.Fatalf("seed member user: %v", err)
	}
	auditService := audit.NewService(auditRepo)
	authService := auth.NewService(cfg.Auth, userRepo, roleService)
	authHandler := auth.NewHandler(authService)
	roleHandler := role.NewHandler(
		roleService,
		auth.RequirePermission(authService, auth.PermissionRolesRead),
		auth.RequirePermission(authService, auth.PermissionRolesWrite),
	)
	auditHandler := audit.NewHandler(auditService, auth.RequirePermission(authService, auth.PermissionAuditRead))
	userHandler := user.NewHandler(
		userService,
		auditService,
		auth.RequireAuth(authService),
		auth.RequirePermission(authService, auth.PermissionUsersRead),
		auth.RequirePermission(authService, auth.PermissionUsersWrite),
	)
	todoHandler := todo.NewHandler(todo.NewHTTPAdapter(todo.NewService(&fakeTodoRepo{})), auth.RequireAuth(authService))
	telemetryProvider := telemetry.NewNoop("pfGoPlus-test")
	edge := bff.New(cfg, authHandler, roleHandler, auditHandler, userHandler, todoHandler, telemetryProvider)
	return httpx.NewRouter(zap.NewNop(), telemetryProvider, edge)
}

type fakeTodoRepo struct{}

func (f *fakeTodoRepo) Create(_ context.Context, item *todo.Todo) error {
	item.ID = 1
	return nil
}

func (f *fakeTodoRepo) List(_ context.Context) ([]todo.Todo, error) {
	return []todo.Todo{}, nil
}

type fakeUserRepo struct {
	items []user.User
}

type fakeRoleRepo struct {
	items []role.Role
}

func (f *fakeUserRepo) Create(_ context.Context, item *user.User) error {
	item.ID = uint(len(f.items) + 1)
	f.items = append(f.items, *item)
	return nil
}

func (f *fakeUserRepo) CountByRole(_ context.Context, role string) (int64, error) {
	var count int64
	for _, item := range f.items {
		if item.Role == role {
			count++
		}
	}
	return count, nil
}

func (f *fakeUserRepo) FindByID(_ context.Context, id uint) (*user.User, error) {
	for i := range f.items {
		if f.items[i].ID == id {
			item := f.items[i]
			return &item, nil
		}
	}
	return nil, nil
}

func (f *fakeUserRepo) FindByUsername(_ context.Context, username string) (*user.User, error) {
	for i := range f.items {
		if f.items[i].Username == username {
			item := f.items[i]
			return &item, nil
		}
	}
	return nil, nil
}

func (f *fakeUserRepo) List(_ context.Context) ([]user.User, error) {
	items := make([]user.User, len(f.items))
	copy(items, f.items)
	return items, nil
}

func (f *fakeUserRepo) Update(_ context.Context, item *user.User) error {
	for i := range f.items {
		if f.items[i].ID == item.ID {
			f.items[i] = *item
			return nil
		}
	}
	return nil
}

func newFakeRoleRepo() *fakeRoleRepo {
	return &fakeRoleRepo{}
}

func (f *fakeRoleRepo) Create(_ context.Context, item *role.Role) error {
	item.ID = uint(len(f.items) + 1)
	if item.Status == "" {
		item.Status = role.StatusActive
	}
	f.items = append(f.items, *item)
	return nil
}

func (f *fakeRoleRepo) Update(_ context.Context, item *role.Role) error {
	for i := range f.items {
		if f.items[i].Name == item.Name {
			f.items[i] = *item
			return nil
		}
	}
	return nil
}

func (f *fakeRoleRepo) FindByName(_ context.Context, name string) (*role.Role, error) {
	for i := range f.items {
		if f.items[i].Name == name {
			item := f.items[i]
			return &item, nil
		}
	}
	return nil, nil
}

func (f *fakeRoleRepo) List(_ context.Context) ([]role.Role, error) {
	items := make([]role.Role, len(f.items))
	copy(items, f.items)
	return items, nil
}

type fakeAuditRepo struct {
	items []audit.Log
}

func (f *fakeAuditRepo) Create(_ context.Context, item *audit.Log) error {
	item.ID = uint(len(f.items) + 1)
	f.items = append(f.items, *item)
	return nil
}

func (f *fakeAuditRepo) List(_ context.Context, query audit.ListQuery) ([]audit.Log, int64, error) {
	filtered := make([]audit.Log, 0, len(f.items))
	for _, item := range f.items {
		if query.ActorUsername != "" && item.ActorUsername != query.ActorUsername {
			continue
		}
		if query.Action != "" && item.Action != query.Action {
			continue
		}
		if query.Resource != "" && item.Resource != query.Resource {
			continue
		}
		if query.Status != "" && item.Status != query.Status {
			continue
		}
		if query.TraceID != "" && item.TraceID != query.TraceID {
			continue
		}
		filtered = append(filtered, item)
	}
	total := int64(len(filtered))
	if query.Offset > len(filtered) {
		return nil, total, nil
	}
	end := query.Offset + query.Limit
	if end > len(filtered) {
		end = len(filtered)
	}
	items := make([]audit.Log, end-query.Offset)
	copy(items, filtered[query.Offset:end])
	return items, total, nil
}
