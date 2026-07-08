package user

import (
	"net/http"

	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
	authz   gin.HandlerFunc
	adminz  gin.HandlerFunc
}

func NewHandler(service *Service, authz gin.HandlerFunc, adminz gin.HandlerFunc) *Handler {
	return &Handler{
		service: service,
		authz:   authz,
		adminz:  adminz,
	}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	users := group.Group("/users")
	users.Use(h.authz)
	users.GET("/me", h.Me)
	users.GET("", h.adminz, h.List)
	users.POST("", h.adminz, h.Create)
}

func (h *Handler) Me(c *gin.Context) {
	userID, ok := c.Get("auth_user_id")
	if !ok {
		_ = c.Error(httpx.Unauthorized("missing user claims", nil))
		return
	}
	httpx.OK(c, gin.H{
		"user": gin.H{
			"id":           userID,
			"username":     c.GetString("auth_username"),
			"display_name": c.GetString("auth_display_name"),
			"role":         c.GetString("auth_role"),
		},
	})
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}

	response := make([]gin.H, 0, len(items))
	for _, item := range items {
		response = append(response, gin.H{
			"id":           item.ID,
			"username":     item.Username,
			"display_name": item.DisplayName,
			"role":         item.Role,
			"status":       item.Status,
			"created_at":   item.CreatedAt,
			"updated_at":   item.UpdatedAt,
		})
	}
	httpx.OK(c, gin.H{"items": response})
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

	httpx.Success(c, http.StatusCreated, "user created", gin.H{
		"user": gin.H{
			"id":           item.ID,
			"username":     item.Username,
			"display_name": item.DisplayName,
			"role":         item.Role,
			"status":       item.Status,
			"created_at":   item.CreatedAt,
			"updated_at":   item.UpdatedAt,
		},
	})
}
