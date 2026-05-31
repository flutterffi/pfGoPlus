package todo

import (
	"net/http"

	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	edge  HTTPEdge
	authz gin.HandlerFunc
}

func NewHandler(edge HTTPEdge, authz gin.HandlerFunc) *Handler {
	return &Handler{edge: edge, authz: authz}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	protected := group.Group("")
	protected.Use(h.authz)
	protected.GET("/todos", h.List)
	protected.POST("/todos", h.Create)
}

func (h *Handler) List(c *gin.Context) {
	response, err := h.edge.List(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	httpx.OK(c, response)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateHTTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(httpx.BadRequest("invalid request body", err))
		return
	}

	response, err := h.edge.Create(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	httpx.Success(c, http.StatusCreated, "todo created", response)
}
