//go:build wireinject

package bootstrap

import (
	"github.com/flutterffi/pfGoPlus/internal/app"
	"github.com/flutterffi/pfGoPlus/internal/config"
	"github.com/google/wire"
)

func InitializeHTTPApp() (*app.HTTPApp, error) {
	wire.Build(
		config.Load,
		NewLogger,
		NewDatabase,
		NewTelemetry,
		MigrateDatabase,
		NewAuthService,
		NewAuthHandler,
		NewTodoRepository,
		NewTodoService,
		NewTodoHandler,
		NewHTTPRouter,
		app.NewHTTPApp,
	)
	return nil, nil
}

func InitializeGRPCApp() (*app.GRPCApp, error) {
	wire.Build(
		config.Load,
		NewLogger,
		NewDatabase,
		NewTelemetry,
		MigrateDatabase,
		NewGRPCServer,
		app.NewGRPCApp,
	)
	return nil, nil
}
