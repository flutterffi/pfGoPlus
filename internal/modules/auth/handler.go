package auth

import (
	"net/http"

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
	authGroup := group.Group("/auth")
	authGroup.POST("/login", h.Login)
	authGroup.GET("/permissions", h.readz, h.ListPermissions)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(httpx.BadRequest("invalid request body", err))
		return
	}

	result, err := h.service.Login(req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	httpx.Success(c, http.StatusOK, "login success", gin.H{"token": result})
}

func (h *Handler) ListPermissions(c *gin.Context) {
	httpx.OK(c, gin.H{
		"items":          Catalog(),
		"groups":         Groups(),
		"role_templates": RoleTemplates(),
	})
}
