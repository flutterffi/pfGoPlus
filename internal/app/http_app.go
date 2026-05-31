package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/flutterffi/pfGoPlus/internal/config"
	"github.com/flutterffi/pfGoPlus/internal/platform/telemetry"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type HTTPApp struct {
	cfg       config.Config
	logger    *zap.Logger
	db        *gorm.DB
	server    *http.Server
	telemetry *telemetry.Provider
	closers   []io.Closer
}

func NewHTTPApp(cfg config.Config, logger *zap.Logger, db *gorm.DB, telemetryProvider *telemetry.Provider, handler http.Handler, closers ...io.Closer) *HTTPApp {
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:           handler,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	return &HTTPApp{
		cfg:       cfg,
		logger:    logger,
		db:        db,
		server:    server,
		telemetry: telemetryProvider,
		closers:   closers,
	}
}

func (a *HTTPApp) Run(ctx context.Context) error {
	serverErr := make(chan error, 1)

	go func() {
		a.logger.Info("http server started", zap.String("addr", a.server.Addr))
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
		return a.cleanup(shutdownCtx)
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("listen and serve: %w", err)
		}
		return nil
	}
}

func (a *HTTPApp) cleanup(ctx context.Context) error {
	if err := closeDB(a.db); err != nil {
		return err
	}
	if a.telemetry != nil {
		if err := a.telemetry.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown telemetry: %w", err)
		}
	}
	for _, closer := range a.closers {
		if closer == nil {
			continue
		}
		if err := closer.Close(); err != nil {
			return fmt.Errorf("close dependency: %w", err)
		}
	}
	a.logger.Info("application stopped")
	return nil
}

func closeDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}
	return nil
}
