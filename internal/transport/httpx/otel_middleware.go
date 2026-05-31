package httpx

import (
	"net/http"
	"time"

	"github.com/flutterffi/pfGoPlus/internal/platform/telemetry"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const otelTraceHeader = "X-Otel-Trace-ID"

func Telemetry(provider *telemetry.Provider) gin.HandlerFunc {
	tracer := provider.Tracer("http-server")
	propagator := provider.Propagator()
	metrics, err := telemetry.NewHTTPMetrics(provider)
	if err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		start := time.Now()
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		spanName := c.Request.Method + " " + c.FullPath()
		if c.FullPath() == "" {
			spanName = c.Request.Method + " " + c.Request.URL.Path
		}

		ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		c.Request = c.Request.WithContext(ctx)
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", c.FullPath()),
			attribute.String("http.target", c.Request.URL.Path),
		)

		spanContext := span.SpanContext()
		if spanContext.HasTraceID() {
			c.Header(otelTraceHeader, spanContext.TraceID().String())
			SetLogger(c, Logger(c).With(
				zap.String("otel_trace_id", spanContext.TraceID().String()),
				zap.String("otel_span_id", spanContext.SpanID().String()),
			))
		}

		c.Next()

		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int("http.response.size", c.Writer.Size()),
		)

		if len(c.Errors) > 0 {
			metrics.Record(c.Request.Context(), c.Request.Method, routeLabel(c), c.Writer.Status(), time.Since(start))
			span.RecordError(c.Errors.Last().Err)
			span.SetStatus(codes.Error, c.Errors.Last().Err.Error())
			return
		}
		if c.Writer.Status() >= http.StatusBadRequest {
			metrics.Record(c.Request.Context(), c.Request.Method, routeLabel(c), c.Writer.Status(), time.Since(start))
			span.SetStatus(codes.Error, http.StatusText(c.Writer.Status()))
			return
		}
		metrics.Record(c.Request.Context(), c.Request.Method, routeLabel(c), c.Writer.Status(), time.Since(start))
		span.SetStatus(codes.Ok, "success")
	}
}

func routeLabel(c *gin.Context) string {
	if c.FullPath() != "" {
		return c.FullPath()
	}
	return c.Request.URL.Path
}
