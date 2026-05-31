package bootstrap

import (
	"context"
	"fmt"
	"io"
	"strings"

	todov1 "github.com/flutterffi/pfGoPlus/api/proto/todo/v1"
	"github.com/flutterffi/pfGoPlus/internal/bff"
	"github.com/flutterffi/pfGoPlus/internal/config"
	"github.com/flutterffi/pfGoPlus/internal/modules/auth"
	"github.com/flutterffi/pfGoPlus/internal/modules/todo"
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

func NewTelemetry(cfg config.Config, log *zap.Logger) (*telemetry.Provider, error) {
	return telemetry.New(cfg.Observability, cfg.App.Name, cfg.App.Env, log)
}

func MigrateDatabase(cfg config.Config, db *gorm.DB) (*gorm.DB, error) {
	if !cfg.Database.AutoMigrate {
		return db, nil
	}
	if err := db.AutoMigrate(&todo.Todo{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	return db, nil
}

func NewAuthService(cfg config.Config) *auth.Service {
	return auth.NewService(cfg.Auth)
}

func NewAuthHandler(service *auth.Service) *auth.Handler {
	return auth.NewHandler(service)
}

func NewTodoRepository(db *gorm.DB) *todo.GormRepository {
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

func NewTodoAPI(cfg config.Config, log *zap.Logger, service *todo.Service) (todo.API, io.Closer, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.TodoBackend.Mode)) {
	case "", "local":
		return service, nil, nil
	case "grpc":
		conn, err := grpcx.Dial(context.Background(), cfg.GRPC.ClientTarget)
		if err != nil {
			return nil, nil, err
		}
		log.Info("todo backend uses grpc client", zap.String("target", cfg.GRPC.ClientTarget))
		return todo.NewGRPCClient(conn), conn, nil
	default:
		return nil, nil, fmt.Errorf("unsupported todo backend mode: %s", cfg.TodoBackend.Mode)
	}
}

func NewBFF(cfg config.Config, authHandler *auth.Handler, todoHandler *todo.Handler, telemetryProvider *telemetry.Provider) *bff.Edge {
	return bff.New(cfg, authHandler, todoHandler, telemetryProvider)
}

func NewHTTPRouter(log *zap.Logger, provider *telemetry.Provider, edge *bff.Edge) *gin.Engine {
	return httpx.NewRouter(log, provider, edge)
}

func NewGRPCServer(log *zap.Logger, provider *telemetry.Provider, todoServer todov1.TodoServiceServer) *grpc.Server {
	return grpcx.NewServer(log, provider, todoServer)
}
