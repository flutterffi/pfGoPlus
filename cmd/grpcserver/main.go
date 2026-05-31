package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/flutterffi/pfGoPlus/internal/bootstrap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := bootstrap.InitializeGRPCApp()
	if err != nil {
		log.Fatalf("bootstrap grpc application: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatalf("run grpc application: %v", err)
	}
}
