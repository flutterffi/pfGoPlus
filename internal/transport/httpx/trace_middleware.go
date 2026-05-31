package httpx

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const traceHeader = "X-Trace-ID"

func Trace(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(traceHeader)
		if traceID == "" {
			traceID = uuid.NewString()
		}

		c.Header(traceHeader, traceID)
		SetTraceID(c, traceID)
		SetLogger(c, log.With(zap.String("trace_id", traceID)))
		c.Next()
	}
}
