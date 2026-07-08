package audit

import (
	"context"
	"strings"

	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
)

type Service struct {
	repo Repository
}

type ListQuery struct {
	ActorUsername string
	Action        string
	Resource      string
	Status        string
	TraceID       string
	Limit         int
	Offset        int
}

type ListResult struct {
	Items  []Log
	Total  int64
	Limit  int
	Offset int
}

type RecordRequest struct {
	ActorID       uint
	ActorUsername string
	Action        string
	Resource      string
	ResourceID    string
	Status        string
	TraceID       string
	Detail        string
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, req RecordRequest) error {
	action := strings.TrimSpace(req.Action)
	resource := strings.TrimSpace(req.Resource)
	status := normalizeStatus(req.Status)
	if action == "" || resource == "" {
		return httpx.BadRequest("audit action and resource are required", nil)
	}

	item := &Log{
		ActorID:       req.ActorID,
		ActorUsername: strings.TrimSpace(req.ActorUsername),
		Action:        action,
		Resource:      resource,
		ResourceID:    strings.TrimSpace(req.ResourceID),
		Status:        status,
		TraceID:       strings.TrimSpace(req.TraceID),
		Detail:        strings.TrimSpace(req.Detail),
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return httpx.Internal("create audit log failed", err)
	}
	return nil
}

func (s *Service) List(ctx context.Context, query ListQuery) (*ListResult, error) {
	query = normalizeListQuery(query)

	items, total, err := s.repo.List(ctx, query)
	if err != nil {
		return nil, httpx.Internal("list audit logs failed", err)
	}
	return &ListResult{
		Items:  items,
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, nil
}

func normalizeStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", StatusSuccess:
		return StatusSuccess
	default:
		return StatusFailure
	}
}

func normalizeListQuery(query ListQuery) ListQuery {
	query.ActorUsername = strings.TrimSpace(query.ActorUsername)
	query.Action = strings.TrimSpace(query.Action)
	query.Resource = strings.TrimSpace(query.Resource)
	query.Status = strings.TrimSpace(query.Status)
	query.TraceID = strings.TrimSpace(query.TraceID)
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Limit > 200 {
		query.Limit = 200
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	return query
}
