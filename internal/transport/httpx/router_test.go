package httpx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flutterffi/pfGoPlus/internal/modules/todo"
	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"go.uber.org/zap"
)

func TestHealthEndpoint(t *testing.T) {
	todoHandler := todo.NewHandler(todo.NewService(&fakeTodoRepo{}))
	router := httpx.NewRouter(zap.NewNop(), todoHandler)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Trace-ID") == "" {
		t.Fatal("expected trace id header")
	}
}

func TestCreateTodoEndpoint(t *testing.T) {
	todoHandler := todo.NewHandler(todo.NewService(&fakeTodoRepo{}))
	router := httpx.NewRouter(zap.NewNop(), todoHandler)

	body, _ := json.Marshal(map[string]string{
		"title":       "Ship API",
		"description": "add basic endpoint",
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/todos", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}
}

type fakeTodoRepo struct{}

func (f *fakeTodoRepo) Create(_ context.Context, item *todo.Todo) error {
	item.ID = 1
	return nil
}

func (f *fakeTodoRepo) List(_ context.Context) ([]todo.Todo, error) {
	return []todo.Todo{}, nil
}
