package todo

import (
	"context"
	"time"

	todov1 "github.com/flutterffi/pfGoPlus/api/proto/todo/v1"
	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"google.golang.org/grpc"
)

type GRPCClient struct {
	client todov1.TodoServiceClient
}

func NewGRPCClient(conn grpc.ClientConnInterface) *GRPCClient {
	return &GRPCClient{
		client: todov1.NewTodoServiceClient(conn),
	}
}

func (c *GRPCClient) List(ctx context.Context) ([]Todo, error) {
	response, err := c.client.ListTodos(ctx, &todov1.ListTodosRequest{})
	if err != nil {
		return nil, httpx.Internal("list todos via grpc failed", err)
	}

	items := make([]Todo, 0, len(response.GetItems()))
	for _, item := range response.GetItems() {
		items = append(items, fromProto(item))
	}
	return items, nil
}

func (c *GRPCClient) Create(ctx context.Context, req CreateRequest) (*Todo, error) {
	response, err := c.client.CreateTodo(ctx, &todov1.CreateTodoRequest{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return nil, httpx.Internal("create todo via grpc failed", err)
	}
	if response.GetItem() == nil {
		return nil, httpx.Internal("create todo via grpc returned empty item", nil)
	}

	item := fromProto(response.GetItem())
	return &item, nil
}

func fromProto(item *todov1.Todo) Todo {
	return Todo{
		ID:          uint(item.GetId()),
		Title:       item.GetTitle(),
		Description: item.GetDescription(),
		Status:      item.GetStatus(),
		CreatedAt:   parseProtoTime(item.GetCreatedAt()),
		UpdatedAt:   parseProtoTime(item.GetUpdatedAt()),
	}
}

func parseProtoTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
