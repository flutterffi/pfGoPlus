package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type GRPCMetrics struct {
	requests metric.Int64Counter
	errors   metric.Int64Counter
	latency  metric.Float64Histogram
}

func NewGRPCMetrics(provider *Provider) (*GRPCMetrics, error) {
	meter := provider.Meter("grpc-metrics")

	requests, err := meter.Int64Counter("grpc.server.requests")
	if err != nil {
		return nil, err
	}
	errors, err := meter.Int64Counter("grpc.server.errors")
	if err != nil {
		return nil, err
	}
	latency, err := meter.Float64Histogram("grpc.server.duration.ms")
	if err != nil {
		return nil, err
	}

	return &GRPCMetrics{
		requests: requests,
		errors:   errors,
		latency:  latency,
	}, nil
}

func (m *GRPCMetrics) Record(ctx context.Context, method string, failed bool, elapsed time.Duration) {
	attrs := metric.WithAttributes(attribute.String("rpc.method", method))
	m.requests.Add(ctx, 1, attrs)
	m.latency.Record(ctx, float64(elapsed.Milliseconds()), attrs)
	if failed {
		m.errors.Add(ctx, 1, attrs)
	}
}
