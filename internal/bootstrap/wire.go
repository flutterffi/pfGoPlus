//go:build wireinject
// +build wireinject

package bootstrap

import (
	"github.com/flutterffi/pfGoPlus/internal/app"
	"github.com/flutterffi/pfGoPlus/internal/config"
	"github.com/google/wire"
)

//go:generate ../../bin/wire

func InitializeHTTPApp() (*app.HTTPApp, error) {
	wire.Build(
		config.Load,
		NewLogger,
		NewReadyDatabase,
		NewTelemetry,
		NewUserRepository,
		NewUserService,
		NewAuthService,
		NewAuthHandler,
		NewUserHandler,
		NewTodoRepository,
		NewTodoService,
		NewTodoBackend,
		NewTodoAPI,
		NewTodoHandler,
		NewBFF,
		NewHTTPRouter,
		NewHTTPHandler,
		NewHTTPAppCleanups,
		app.NewHTTPApp,
	)
	return nil, nil
}

func InitializeGRPCApp() (*app.GRPCApp, error) {
	wire.Build(
		config.Load,
		NewLogger,
		NewReadyDatabase,
		NewTelemetry,
		NewTodoRepository,
		NewTodoService,
		NewTodoGRPCService,
		NewGRPCServer,
		app.NewGRPCApp,
	)
	return nil, nil
}
