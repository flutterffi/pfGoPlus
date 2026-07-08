package audit

import (
	"strconv"

	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
	adminz  gin.HandlerFunc
}

func NewHandler(service *Service, adminz gin.HandlerFunc) *Handler {
	return &Handler{
		service: service,
		adminz:  adminz,
	}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	logs := group.Group("/audit/logs")
	logs.Use(h.adminz)
	logs.GET("", h.List)
}

func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	items, err := h.service.List(c.Request.Context(), limit)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response := make([]gin.H, 0, len(items))
	for _, item := range items {
		response = append(response, gin.H{
			"id":             item.ID,
			"actor_id":       item.ActorID,
			"actor_username": item.ActorUsername,
			"action":         item.Action,
			"resource":       item.Resource,
			"resource_id":    item.ResourceID,
			"status":         item.Status,
			"trace_id":       item.TraceID,
			"detail":         item.Detail,
			"created_at":     item.CreatedAt,
		})
	}

	httpx.OK(c, gin.H{"items": response})
}
