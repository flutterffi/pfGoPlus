package audit

import (
	"context"
	"strings"

	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
)

type Service struct {
	repo Repository
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

func (s *Service) List(ctx context.Context, limit int) ([]Log, error) {
	items, err := s.repo.List(ctx, limit)
	if err != nil {
		return nil, httpx.Internal("list audit logs failed", err)
	}
	return items, nil
}

func normalizeStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", StatusSuccess:
		return StatusSuccess
	default:
		return StatusFailure
	}
}
