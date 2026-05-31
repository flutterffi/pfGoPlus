package todo

import (
	"context"
	"time"

	todov1 "github.com/flutterffi/pfGoPlus/api/proto/todo/v1"
)

type GRPCService struct {
	todov1.UnimplementedTodoServiceServer
	service *Service
}

func NewGRPCService(service *Service) *GRPCService {
	return &GRPCService{service: service}
}

func (s *GRPCService) ListTodos(ctx context.Context, _ *todov1.ListTodosRequest) (*todov1.ListTodosResponse, error) {
	items, err := s.service.List(ctx)
	if err != nil {
		return nil, err
	}

	response := &todov1.ListTodosResponse{
		Items: make([]*todov1.Todo, 0, len(items)),
	}
	for _, item := range items {
		response.Items = append(response.Items, mapTodo(item))
	}
	return response, nil
}

func (s *GRPCService) CreateTodo(ctx context.Context, req *todov1.CreateTodoRequest) (*todov1.CreateTodoResponse, error) {
	item, err := s.service.Create(ctx, CreateRequest{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	return &todov1.CreateTodoResponse{Item: mapTodo(*item)}, nil
}

func mapTodo(item Todo) *todov1.Todo {
	return &todov1.Todo{
		ID:          uint64(item.ID),
		Title:       item.Title,
		Description: item.Description,
		Status:      item.Status,
		CreatedAt:   formatTime(item.CreatedAt),
		UpdatedAt:   formatTime(item.UpdatedAt),
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
