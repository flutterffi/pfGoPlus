package todo

import (
	"context"
	"net"
	"testing"

	"github.com/flutterffi/pfGoPlus/internal/platform/telemetry"
	"github.com/flutterffi/pfGoPlus/internal/transport/grpcx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestGRPCClientRoundTrip(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpcx.NewServer(zap.NewNop(), telemetry.NewNoop("pfGoPlus-test"), NewGRPCService(NewService(&stubRepository{})))
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
		t.Fatalf("dial grpc server: %v", err)
	}
	defer conn.Close()

	client := NewGRPCClient(conn)
	item, err := client.Create(context.Background(), CreateRequest{
		Title:       "Ship adapter",
		Description: "verify grpc client mode",
	})
	if err != nil {
		t.Fatalf("create via grpc client: %v", err)
	}
	if item.Title != "Ship adapter" {
		t.Fatalf("unexpected title: %s", item.Title)
	}

	items, err := client.List(context.Background())
	if err != nil {
		t.Fatalf("list via grpc client: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}
