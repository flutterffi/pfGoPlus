package httpx

import (
	"net/http"

	"github.com/flutterffi/pfGoPlus/internal/platform/telemetry"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RouterComposer interface {
	Compose(router *gin.Engine)
}

func NewRouter(log *zap.Logger, provider *telemetry.Provider, composer RouterComposer) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(Trace(log))
	router.Use(Telemetry(provider))
	router.Use(RequestLogger())
	router.Use(Recovery())
	router.Use(ErrorHandler())

	if composer != nil {
		composer.Compose(router)
	}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":     404000,
			"message":  "resource not found",
			"trace_id": TraceID(c),
		})
	})

	return router
}
