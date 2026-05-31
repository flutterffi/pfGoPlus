package todo

import (
	"context"
	"testing"
	"time"
)

type stubEdgeAPI struct {
	items []Todo
}

func (s *stubEdgeAPI) Create(_ context.Context, req CreateRequest) (*Todo, error) {
	item := &Todo{
		ID:          uint(len(s.items) + 1),
		Title:       req.Title,
		Description: req.Description,
		Status:      "pending",
		CreatedAt:   time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}
	s.items = append(s.items, *item)
	return item, nil
}

func (s *stubEdgeAPI) List(_ context.Context) ([]Todo, error) {
	return s.items, nil
}

func TestHTTPAdapterCreateMapsResponse(t *testing.T) {
	adapter := NewHTTPAdapter(&stubEdgeAPI{})

	response, err := adapter.Create(context.Background(), CreateHTTPRequest{
		Title:       "Ship edge adapter",
		Description: "stabilize http contract",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response.Item.ID != 1 {
		t.Fatalf("expected id 1, got %d", response.Item.ID)
	}
	if response.Item.CreatedAt != "2026-01-02T03:04:05Z" {
		t.Fatalf("unexpected created_at: %s", response.Item.CreatedAt)
	}
}
