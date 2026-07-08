package audit

import (
	"context"
	"testing"
)

type stubRepository struct {
	items []Log
}

func (s *stubRepository) Create(_ context.Context, item *Log) error {
	item.ID = uint(len(s.items) + 1)
	s.items = append(s.items, *item)
	return nil
}

func (s *stubRepository) List(_ context.Context, limit int) ([]Log, error) {
	if limit <= 0 || limit > len(s.items) {
		limit = len(s.items)
	}
	items := make([]Log, limit)
	copy(items, s.items[:limit])
	return items, nil
}

func TestRecordSuccess(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)

	err := service.Record(context.Background(), RecordRequest{
		ActorID:       1,
		ActorUsername: "admin",
		Action:        "user.create",
		Resource:      "user",
		ResourceID:    "2",
		Status:        StatusSuccess,
		TraceID:       "trace-1",
		Detail:        "created user alice",
	})
	if err != nil {
		t.Fatalf("record audit log: %v", err)
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(repo.items))
	}
	if repo.items[0].Action != "user.create" {
		t.Fatalf("unexpected action: %s", repo.items[0].Action)
	}
}
