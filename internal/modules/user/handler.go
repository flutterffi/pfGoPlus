package user

import (
	"net/http"
	"strconv"

	"github.com/flutterffi/pfGoPlus/internal/modules/audit"
	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
	audit   *audit.Service
	authz   gin.HandlerFunc
	adminz  gin.HandlerFunc
}

func NewHandler(service *Service, auditService *audit.Service, authz gin.HandlerFunc, adminz gin.HandlerFunc) *Handler {
	return &Handler{
		service: service,
		audit:   auditService,
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
	users.PATCH("/:id", h.adminz, h.Update)
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
		copyItem := item
		response = append(response, presentUser(&copyItem))
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
		"user": presentUser(item),
	})
	h.recordAudit(c, "user.create", "user", strconv.FormatUint(uint64(item.ID), 10), audit.StatusSuccess, "created user "+item.Username)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		_ = c.Error(httpx.BadRequest("invalid user id", err))
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(httpx.BadRequest("invalid request body", err))
		return
	}

	actorID, _ := c.Get("auth_user_id")
	currentActorID, _ := actorID.(uint)

	item, updateErr := h.service.Update(c.Request.Context(), uint(id), req, currentActorID)
	if updateErr != nil {
		_ = c.Error(updateErr)
		return
	}

	httpx.OK(c, gin.H{
		"user": presentUser(item),
	})
	h.recordAudit(c, "user.update", "user", strconv.FormatUint(uint64(item.ID), 10), audit.StatusSuccess, "updated user "+item.Username)
}

func presentUser(item *User) gin.H {
	return gin.H{
		"id":           item.ID,
		"username":     item.Username,
		"display_name": item.DisplayName,
		"role":         item.Role,
		"status":       item.Status,
		"created_at":   item.CreatedAt,
		"updated_at":   item.UpdatedAt,
	}
}

func (h *Handler) recordAudit(c *gin.Context, action, resource, resourceID, status, detail string) {
	if h.audit == nil {
		return
	}
	actorID, ok := c.Get("auth_user_id")
	if !ok {
		return
	}
	currentActorID, _ := actorID.(uint)
	if err := h.audit.Record(c.Request.Context(), audit.RecordRequest{
		ActorID:       currentActorID,
		ActorUsername: c.GetString("auth_username"),
		Action:        action,
		Resource:      resource,
		ResourceID:    resourceID,
		Status:        status,
		TraceID:       httpx.TraceID(c),
		Detail:        detail,
	}); err != nil {
		httpx.Logger(c).Error("record audit log failed")
	}
}
