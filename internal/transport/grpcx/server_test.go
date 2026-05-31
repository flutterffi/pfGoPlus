package grpcx

import (
	"context"
	"net"
	"testing"

	"github.com/flutterffi/pfGoPlus/internal/platform/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"
)

func TestHealthCheck(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := NewServer(zap.NewNop(), telemetry.NewNoop("pfGoPlus-test"))
	defer server.Stop()

	go func() {
		if err := server.Serve(listener); err != nil {
			t.Logf("grpc server stopped: %v", err)
		}
	}()

	conn, err := grpc.DialContext(
		context.Background(),
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("create grpc client: %v", err)
	}
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)
	response, err := client.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	if response.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("expected serving status, got %s", response.GetStatus().String())
	}
}
