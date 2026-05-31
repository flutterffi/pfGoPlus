package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/flutterffi/pfGoPlus/internal/config"
	"github.com/flutterffi/pfGoPlus/internal/modules/todo"
	"github.com/flutterffi/pfGoPlus/internal/platform/database"
	platformlogger "github.com/flutterffi/pfGoPlus/internal/platform/logger"
	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	cfg    config.Config
	logger *zap.Logger
	db     *gorm.DB
	server *http.Server
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	log, err := platformlogger.New(cfg.Logger, cfg.App.Name, cfg.App.Env)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	db, err := database.New(cfg.Database, log)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	if cfg.Database.AutoMigrate {
		if err := db.AutoMigrate(&todo.Todo{}); err != nil {
			return nil, fmt.Errorf("auto migrate: %w", err)
		}
	}

	todoRepo := todo.NewRepository(db)
	todoService := todo.NewService(todoRepo)
	todoHandler := todo.NewHandler(todoService)
	router := httpx.NewRouter(log, todoHandler)

	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
	}

	return &App{
		cfg:    cfg,
		logger: log,
		db:     db,
		server: server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	serverErr := make(chan error, 1)

	go func() {
		a.logger.Info("http server started",
			zap.String("addr", a.server.Addr),
		)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		a.logger.Info("shutdown signal received")
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}

		if err := a.closeDB(); err != nil {
			return err
		}

		a.logger.Info("application stopped")
		return nil
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("listen and serve: %w", err)
		}
		return nil
	}
}

func (a *App) closeDB() error {
	sqlDB, err := a.db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}
	return nil
}
