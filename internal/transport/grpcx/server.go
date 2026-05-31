package grpcx

import (
	"context"
	"time"

	todov1 "github.com/flutterffi/pfGoPlus/api/proto/todo/v1"
	"github.com/flutterffi/pfGoPlus/internal/platform/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func NewServer(log *zap.Logger, provider *telemetry.Provider, todoServer todov1.TodoServiceServer) *grpc.Server {
	metrics, err := telemetry.NewGRPCMetrics(provider)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithTracerProvider(provider.TracerProvider()),
		)),
		grpc.ChainUnaryInterceptor(LoggingUnaryInterceptor(log, metrics)),
	)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	todov1.RegisterTodoServiceServer(server, todoServer)
	reflection.Register(server)

	return server
}

func LoggingUnaryInterceptor(log *zap.Logger, metrics *telemetry.GRPCMetrics) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)

		spanContext := trace.SpanFromContext(ctx).SpanContext()
		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("latency", time.Since(start)),
		}
		if spanContext.HasTraceID() {
			fields = append(fields,
				zap.String("otel_trace_id", spanContext.TraceID().String()),
				zap.String("otel_span_id", spanContext.SpanID().String()),
			)
		}
		if err != nil {
			metrics.Record(ctx, info.FullMethod, true, time.Since(start))
			fields = append(fields, zap.Error(err))
			log.Error("grpc request failed", fields...)
			return resp, err
		}

		metrics.Record(ctx, info.FullMethod, false, time.Since(start))
		log.Info("grpc request completed", fields...)
		return resp, nil
	}
}
