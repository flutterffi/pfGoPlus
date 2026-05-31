package todo

import (
	"context"
	"testing"
)

type stubRepository struct {
	items []Todo
}

func (s *stubRepository) Create(_ context.Context, item *Todo) error {
	item.ID = uint(len(s.items) + 1)
	s.items = append(s.items, *item)
	return nil
}

func (s *stubRepository) List(_ context.Context) ([]Todo, error) {
	return s.items, nil
}

func TestCreateTodoRequiresTitle(t *testing.T) {
	service := NewService(&stubRepository{})

	_, err := service.Create(context.Background(), CreateRequest{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreateTodoSuccess(t *testing.T) {
	service := NewService(&stubRepository{})

	item, err := service.Create(context.Background(), CreateRequest{
		Title:       "Ship scaffold",
		Description: "build monolith base",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if item.Title != "Ship scaffold" {
		t.Fatalf("unexpected title: %s", item.Title)
	}
	if item.Status != "pending" {
		t.Fatalf("unexpected status: %s", item.Status)
	}
}
