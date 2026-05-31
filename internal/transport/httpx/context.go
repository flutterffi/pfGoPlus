package httpx

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	traceIDKey = "trace_id"
	loggerKey  = "logger"
)

func TraceID(c *gin.Context) string {
	if value, ok := c.Get(traceIDKey); ok {
		if traceID, ok := value.(string); ok {
			return traceID
		}
	}
	return ""
}

func SetTraceID(c *gin.Context, traceID string) {
	c.Set(traceIDKey, traceID)
}

func SetLogger(c *gin.Context, logger *zap.Logger) {
	c.Set(loggerKey, logger)
}

func Logger(c *gin.Context) *zap.Logger {
	if value, ok := c.Get(loggerKey); ok {
		if logger, ok := value.(*zap.Logger); ok {
			return logger
		}
	}
	return zap.NewNop()
}
