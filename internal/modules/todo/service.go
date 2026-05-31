package todo

import (
	"context"
	"strings"

	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
)

type Service struct {
	repo Repository
}

type CreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Todo, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, httpx.BadRequest("title is required", nil)
	}

	item := &Todo{
		Title:       title,
		Description: strings.TrimSpace(req.Description),
		Status:      "pending",
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, httpx.Internal("create todo failed", err)
	}
	return item, nil
}

func (s *Service) List(ctx context.Context) ([]Todo, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, httpx.Internal("list todos failed", err)
	}
	return items, nil
}
