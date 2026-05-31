package grpcx

import (
	"context"
	"net"
	"testing"

	todov1 "github.com/flutterffi/pfGoPlus/api/proto/todo/v1"
	"github.com/flutterffi/pfGoPlus/internal/modules/todo"
	"github.com/flutterffi/pfGoPlus/internal/platform/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"
)

func TestHealthCheck(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := NewServer(zap.NewNop(), telemetry.NewNoop("pfGoPlus-test"), todo.NewGRPCService(todo.NewService(&fakeTodoRepo{})))
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

func TestTodoServiceRoundTrip(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := NewServer(zap.NewNop(), telemetry.NewNoop("pfGoPlus-test"), todo.NewGRPCService(todo.NewService(&fakeTodoRepo{})))
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

	client := todov1.NewTodoServiceClient(conn)
	createResp, err := client.CreateTodo(context.Background(), &todov1.CreateTodoRequest{
		Title:       "Ship gRPC todo",
		Description: "validate business rpc",
	})
	if err != nil {
		t.Fatalf("create todo rpc failed: %v", err)
	}
	if createResp.Item == nil || createResp.Item.Title != "Ship gRPC todo" {
		t.Fatal("expected created todo item in response")
	}

	listResp, err := client.ListTodos(context.Background(), &todov1.ListTodosRequest{})
	if err != nil {
		t.Fatalf("list todos rpc failed: %v", err)
	}
	if len(listResp.Items) == 0 {
		t.Fatal("expected at least one todo item")
	}
}

type fakeTodoRepo struct {
	items []todo.Todo
}

func (f *fakeTodoRepo) Create(_ context.Context, item *todo.Todo) error {
	item.ID = uint(len(f.items) + 1)
	f.items = append(f.items, *item)
	return nil
}

func (f *fakeTodoRepo) List(_ context.Context) ([]todo.Todo, error) {
	items := make([]todo.Todo, len(f.items))
	copy(items, f.items)
	return items, nil
}
