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

func (s *stubRepository) List(_ context.Context, query ListQuery) ([]Log, int64, error) {
	filtered := make([]Log, 0, len(s.items))
	for _, item := range s.items {
		if query.ActorUsername != "" && item.ActorUsername != query.ActorUsername {
			continue
		}
		if query.Action != "" && item.Action != query.Action {
			continue
		}
		if query.Resource != "" && item.Resource != query.Resource {
			continue
		}
		if query.Status != "" && item.Status != query.Status {
			continue
		}
		if query.TraceID != "" && item.TraceID != query.TraceID {
			continue
		}
		filtered = append(filtered, item)
	}
	total := int64(len(filtered))
	if query.Offset > len(filtered) {
		return nil, total, nil
	}
	end := query.Offset + query.Limit
	if end > len(filtered) {
		end = len(filtered)
	}
	items := make([]Log, end-query.Offset)
	copy(items, filtered[query.Offset:end])
	return items, total, nil
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

func TestListSupportsFilteringAndPagination(t *testing.T) {
	repo := &stubRepository{
		items: []Log{
			{ID: 1, ActorUsername: "admin", Action: "user.create", Resource: "user", Status: StatusSuccess, TraceID: "trace-1"},
			{ID: 2, ActorUsername: "admin", Action: "user.update", Resource: "user", Status: StatusSuccess, TraceID: "trace-2"},
			{ID: 3, ActorUsername: "alice", Action: "todo.create", Resource: "todo", Status: StatusFailure, TraceID: "trace-3"},
		},
	}
	service := NewService(repo)

	result, err := service.List(context.Background(), ListQuery{
		ActorUsername: "admin",
		Resource:      "user",
		Limit:         1,
		Offset:        1,
	})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("expected total 2, got %d", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 paged item, got %d", len(result.Items))
	}
	if result.Items[0].Action != "user.update" {
		t.Fatalf("unexpected action: %s", result.Items[0].Action)
	}
}
