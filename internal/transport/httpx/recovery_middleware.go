package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				Logger(c).Error("panic recovered", zap.Any("panic", recovered))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":     500001,
					"message":  "internal server error",
					"trace_id": TraceID(c),
				})
			}
		}()

		c.Next()
	}
}
