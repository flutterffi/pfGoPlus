package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/flutterffi/pfGoPlus/internal/config"
	"github.com/flutterffi/pfGoPlus/internal/platform/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type GRPCApp struct {
	cfg       config.Config
	logger    *zap.Logger
	db        *gorm.DB
	server    *grpc.Server
	telemetry *telemetry.Provider
}

func NewGRPCApp(cfg config.Config, logger *zap.Logger, db *gorm.DB, telemetryProvider *telemetry.Provider, server *grpc.Server) *GRPCApp {
	return &GRPCApp{
		cfg:       cfg,
		logger:    logger,
		db:        db,
		server:    server,
		telemetry: telemetryProvider,
	}
}

func (a *GRPCApp) Run(ctx context.Context) error {
	address := fmt.Sprintf("%s:%d", a.cfg.GRPC.Host, a.cfg.GRPC.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	serverErr := make(chan error, 1)
	go func() {
		a.logger.Info("grpc server started", zap.String("addr", address))
		if err := a.server.Serve(listener); err != nil {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		a.server.GracefulStop()
		return a.cleanup(stopCtx)
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("serve grpc: %w", err)
		}
		return nil
	}
}

func (a *GRPCApp) cleanup(ctx context.Context) error {
	if err := closeDB(a.db); err != nil {
		return err
	}
	if a.telemetry != nil {
		if err := a.telemetry.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown telemetry: %w", err)
		}
	}
	a.logger.Info("grpc application stopped")
	return nil
}
