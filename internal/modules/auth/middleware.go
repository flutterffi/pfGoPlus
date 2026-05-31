package auth

import (
	"strings"

	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"github.com/gin-gonic/gin"
)

const authUserKey = "auth_user"

func RequireAuth(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if header == "" {
			_ = c.Error(httpx.Unauthorized("missing authorization header", nil))
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			_ = c.Error(httpx.Unauthorized("invalid authorization header", nil))
			c.Abort()
			return
		}

		claims, err := service.ParseToken(strings.TrimSpace(parts[1]))
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}

		c.Set(authUserKey, claims.Username)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) string {
	value, ok := c.Get(authUserKey)
	if !ok {
		return ""
	}
	username, _ := value.(string)
	return username
}
