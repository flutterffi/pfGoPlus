package role

import (
	"encoding/json"

	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
	readz   gin.HandlerFunc
}

func NewHandler(service *Service, readz gin.HandlerFunc) *Handler {
	return &Handler{
		service: service,
		readz:   readz,
	}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	roles := group.Group("/roles")
	roles.Use(h.readz)
	roles.GET("", h.List)
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}

	response := make([]gin.H, 0, len(items))
	for _, item := range items {
		var permissions []string
		_ = json.Unmarshal([]byte(item.Permissions), &permissions)
		response = append(response, gin.H{
			"id":           item.ID,
			"name":         item.Name,
			"display_name": item.DisplayName,
			"permissions":  permissions,
			"created_at":   item.CreatedAt,
			"updated_at":   item.UpdatedAt,
		})
	}
	httpx.OK(c, gin.H{"items": response})
}
