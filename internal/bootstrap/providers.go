package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	todov1 "github.com/flutterffi/pfGoPlus/api/proto/todo/v1"
	"github.com/flutterffi/pfGoPlus/internal/app"
	"github.com/flutterffi/pfGoPlus/internal/bff"
	"github.com/flutterffi/pfGoPlus/internal/config"
	"github.com/flutterffi/pfGoPlus/internal/modules/auth"
	"github.com/flutterffi/pfGoPlus/internal/modules/todo"
	"github.com/flutterffi/pfGoPlus/internal/modules/user"
	"github.com/flutterffi/pfGoPlus/internal/platform/database"
	"github.com/flutterffi/pfGoPlus/internal/platform/logger"
	"github.com/flutterffi/pfGoPlus/internal/platform/telemetry"
	"github.com/flutterffi/pfGoPlus/internal/transport/grpcx"
	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

func NewLogger(cfg config.Config) (*zap.Logger, error) {
	return logger.New(cfg.Logger, cfg.App.Name, cfg.App.Env)
}

func NewDatabase(cfg config.Config, log *zap.Logger) (*gorm.DB, error) {
	return database.New(cfg.Database, log)
}

func NewReadyDatabase(cfg config.Config, log *zap.Logger) (*gorm.DB, error) {
	db, err := NewDatabase(cfg, log)
	if err != nil {
		return nil, err
	}
	return MigrateDatabase(cfg, db)
}

func NewTelemetry(cfg config.Config, log *zap.Logger) (*telemetry.Provider, error) {
	return telemetry.New(cfg.Observability, cfg.App.Name, cfg.App.Env, log)
}

func MigrateDatabase(cfg config.Config, db *gorm.DB) (*gorm.DB, error) {
	if !cfg.Database.AutoMigrate {
		return db, nil
	}
	if err := db.AutoMigrate(&user.User{}, &todo.Todo{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	return db, nil
}

func NewUserRepository(db *gorm.DB) user.Repository {
	return user.NewRepository(db)
}

func NewUserService(cfg config.Config, repo user.Repository) (*user.Service, error) {
	return user.NewService(cfg.Auth, repo)
}

func NewAuthService(cfg config.Config, repo user.Repository) *auth.Service {
	return auth.NewService(cfg.Auth, repo)
}

func NewAuthHandler(service *auth.Service) *auth.Handler {
	return auth.NewHandler(service)
}

func NewUserHandler(service *user.Service, authService *auth.Service) *user.Handler {
	return user.NewHandler(service, auth.RequireAuth(authService), auth.RequireRole(authService, user.RoleAdmin))
}

func NewTodoRepository(db *gorm.DB) todo.Repository {
	return todo.NewRepository(db)
}

func NewTodoService(repo todo.Repository) *todo.Service {
	return todo.NewService(repo)
}

func NewTodoHandler(service todo.API, authService *auth.Service) *todo.Handler {
	return todo.NewHandler(todo.NewHTTPAdapter(service), auth.RequireAuth(authService))
}

func NewTodoGRPCService(service *todo.Service) todov1.TodoServiceServer {
	return todo.NewGRPCService(service)
}

type TodoBackend struct {
	API     todo.API
	Cleanup app.Cleanup
}

func NewTodoBackend(cfg config.Config, log *zap.Logger, service *todo.Service) (*TodoBackend, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.TodoBackend.Mode)) {
	case "", "local":
		return &TodoBackend{API: service}, nil
	case "grpc":
		conn, err := grpcx.Dial(context.Background(), cfg.GRPC.ClientTarget)
		if err != nil {
			return nil, err
		}
		log.Info("todo backend uses grpc client", zap.String("target", cfg.GRPC.ClientTarget))
		return &TodoBackend{
			API: todo.NewGRPCClient(conn),
			Cleanup: func() {
				_ = conn.Close()
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported todo backend mode: %s", cfg.TodoBackend.Mode)
	}
}

func NewTodoAPI(backend *TodoBackend) todo.API {
	return backend.API
}

func NewBFF(cfg config.Config, authHandler *auth.Handler, userHandler *user.Handler, todoHandler *todo.Handler, telemetryProvider *telemetry.Provider) *bff.Edge {
	return bff.New(cfg, authHandler, userHandler, todoHandler, telemetryProvider)
}

func NewHTTPRouter(log *zap.Logger, provider *telemetry.Provider, edge *bff.Edge) *gin.Engine {
	return httpx.NewRouter(log, provider, edge)
}

func NewHTTPHandler(engine *gin.Engine) http.Handler {
	return engine
}

func NewHTTPAppCleanups(backend *TodoBackend) []app.Cleanup {
	if backend == nil || backend.Cleanup == nil {
		return nil
	}
	return []app.Cleanup{backend.Cleanup}
}

func NewGRPCServer(log *zap.Logger, provider *telemetry.Provider, todoServer todov1.TodoServiceServer) *grpc.Server {
	return grpcx.NewServer(log, provider, todoServer)
}
