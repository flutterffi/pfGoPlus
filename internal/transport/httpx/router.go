package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RouteRegistrar interface {
	RegisterRoutes(group *gin.RouterGroup)
}

func NewRouter(log *zap.Logger, registrars ...RouteRegistrar) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(Trace(log))
	router.Use(RequestLogger())
	router.Use(Recovery())
	router.Use(ErrorHandler())

	router.GET("/health", func(c *gin.Context) {
		OK(c, gin.H{
			"status": "ok",
		})
	})

	v1 := router.Group("/api/v1")
	for _, registrar := range registrars {
		registrar.RegisterRoutes(v1)
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
