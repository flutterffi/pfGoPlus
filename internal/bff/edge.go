package bff

import (
	"github.com/flutterffi/pfGoPlus/internal/config"
	"github.com/flutterffi/pfGoPlus/internal/modules/auth"
	"github.com/flutterffi/pfGoPlus/internal/modules/todo"
	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"github.com/gin-gonic/gin"
)

type Edge struct {
	cfg         config.Config
	authHandler *auth.Handler
	todoHandler *todo.Handler
}

func New(cfg config.Config, authHandler *auth.Handler, todoHandler *todo.Handler) *Edge {
	return &Edge{
		cfg:         cfg,
		authHandler: authHandler,
		todoHandler: todoHandler,
	}
}

func (e *Edge) Compose(router *gin.Engine) {
	router.GET("/health", e.Health)
	router.GET("/ready", e.Ready)

	v1 := router.Group("/api/v1")
	e.authHandler.RegisterRoutes(v1)
	e.todoHandler.RegisterRoutes(v1)
	v1.GET("/meta", e.Meta)
}

func (e *Edge) Health(c *gin.Context) {
	httpx.OK(c, gin.H{
		"status": "ok",
		"app":    e.cfg.App.Name,
	})
}

func (e *Edge) Ready(c *gin.Context) {
	httpx.OK(c, gin.H{
		"status":  "ready",
		"app":     e.cfg.App.Name,
		"version": e.cfg.Observability.ServiceVersion,
	})
}

func (e *Edge) Meta(c *gin.Context) {
	httpx.OK(c, gin.H{
		"app":          e.cfg.App.Name,
		"env":          e.cfg.App.Env,
		"version":      e.cfg.Observability.ServiceVersion,
		"todo_backend": e.cfg.TodoBackend.Mode,
		"grpc_target":  e.cfg.GRPC.ClientTarget,
	})
}
