package todo

import (
	"net/http"

	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
	authz   gin.HandlerFunc
}

func NewHandler(service *Service, authz gin.HandlerFunc) *Handler {
	return &Handler{service: service, authz: authz}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	protected := group.Group("")
	protected.Use(h.authz)
	protected.GET("/todos", h.List)
	protected.POST("/todos", h.Create)
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	httpx.OK(c, gin.H{"items": items})
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(httpx.BadRequest("invalid request body", err))
		return
	}

	item, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	httpx.Success(c, http.StatusCreated, "todo created", gin.H{"item": item})
}
