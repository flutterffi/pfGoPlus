package auth

import (
	"strings"

	"github.com/flutterffi/pfGoPlus/internal/transport/httpx"
	"github.com/gin-gonic/gin"
)

const (
	authUserKey        = "auth_user"
	authUserIDKey      = "auth_user_id"
	authUsernameKey    = "auth_username"
	authDisplayNameKey = "auth_display_name"
	authRoleKey        = "auth_role"
	authPermissionsKey = "auth_permissions"
)

func RequireAuth(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := authenticate(service, c)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}

		c.Set(authUserKey, claims)
		c.Set(authUserIDKey, claims.UserID)
		c.Set(authUsernameKey, claims.Username)
		c.Set(authDisplayNameKey, claims.DisplayName)
		c.Set(authRoleKey, claims.Role)
		c.Set(authPermissionsKey, claims.Permissions)
		c.Next()
	}
}

func RequireRole(service *Service, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := authenticate(service, c)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}
		c.Set(authUserKey, claims)
		c.Set(authUserIDKey, claims.UserID)
		c.Set(authUsernameKey, claims.Username)
		c.Set(authDisplayNameKey, claims.DisplayName)
		c.Set(authRoleKey, claims.Role)
		c.Set(authPermissionsKey, claims.Permissions)
		for _, role := range roles {
			if strings.EqualFold(claims.Role, role) {
				c.Next()
				return
			}
		}

		_ = c.Error(httpx.Forbidden("insufficient permissions", nil))
		c.Abort()
	}
}

func RequirePermission(service *Service, permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := authenticate(service, c)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}
		c.Set(authUserKey, claims)
		c.Set(authUserIDKey, claims.UserID)
		c.Set(authUsernameKey, claims.Username)
		c.Set(authDisplayNameKey, claims.DisplayName)
		c.Set(authRoleKey, claims.Role)
		c.Set(authPermissionsKey, claims.Permissions)
		for _, permission := range permissions {
			if hasPermission(claims.Permissions, permission) {
				c.Next()
				return
			}
		}

		_ = c.Error(httpx.Forbidden("insufficient permissions", nil))
		c.Abort()
	}
}

func CurrentClaims(c *gin.Context) *Claims {
	value, ok := c.Get(authUserKey)
	if !ok {
		return nil
	}
	claims, _ := value.(*Claims)
	return claims
}

func CurrentUser(c *gin.Context) string {
	claims := CurrentClaims(c)
	if claims == nil {
		return ""
	}
	return claims.Username
}

func CurrentPermissions(c *gin.Context) []string {
	value, ok := c.Get(authPermissionsKey)
	if !ok {
		return nil
	}
	permissions, _ := value.([]string)
	return permissions
}

func authenticate(service *Service, c *gin.Context) (*Claims, error) {
	header := strings.TrimSpace(c.GetHeader("Authorization"))
	if header == "" {
		return nil, httpx.Unauthorized("missing authorization header", nil)
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, httpx.Unauthorized("invalid authorization header", nil)
	}

	return service.ParseToken(strings.TrimSpace(parts[1]))
}
