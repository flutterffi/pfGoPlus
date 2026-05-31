package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type HTTPMetrics struct {
	requests metric.Int64Counter
	errors   metric.Int64Counter
	latency  metric.Float64Histogram
}

func NewHTTPMetrics(provider *Provider) (*HTTPMetrics, error) {
	meter := provider.Meter("http-metrics")

	requests, err := meter.Int64Counter("http.server.requests")
	if err != nil {
		return nil, err
	}
	errors, err := meter.Int64Counter("http.server.errors")
	if err != nil {
		return nil, err
	}
	latency, err := meter.Float64Histogram("http.server.duration.ms")
	if err != nil {
		return nil, err
	}

	return &HTTPMetrics{
		requests: requests,
		errors:   errors,
		latency:  latency,
	}, nil
}

func (m *HTTPMetrics) Record(ctx context.Context, method, route string, status int, elapsed time.Duration) {
	attrs := metric.WithAttributes(
		attribute.String("http.method", method),
		attribute.String("http.route", route),
		attribute.Int("http.status_code", status),
	)

	m.requests.Add(ctx, 1, attrs)
	m.latency.Record(ctx, float64(elapsed.Milliseconds()), attrs)
	if status >= 400 {
		m.errors.Add(ctx, 1, attrs)
	}
}
