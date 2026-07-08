package telemetry

import (
	"context"
	"strings"
	"testing"

	"github.com/flutterffi/pfGoPlus/internal/config"
	"go.uber.org/zap"
)

func TestNewSupportsOTLPExporter(t *testing.T) {
	provider, err := New(config.ObservabilityConfig{
		Enabled:        true,
		Exporter:       "otlp",
		OTLPEndpoint:   "127.0.0.1:4317",
		OTLPInsecure:   true,
		ServiceVersion: "test-version",
	}, "pfGoPlus-test", "test", zap.NewNop())
	if err != nil {
		t.Fatalf("expected otlp exporter to initialize, got %v", err)
	}
	if provider == nil {
		t.Fatal("expected provider")
	}
	if provider.MetricsHandler() != nil {
		t.Fatal("expected no metrics handler for otlp exporter")
	}
	if err := provider.Shutdown(context.Background()); err != nil &&
		!strings.Contains(err.Error(), "connection refused") &&
		!strings.Contains(err.Error(), "operation not permitted") {
		t.Fatalf("shutdown telemetry provider: %v", err)
	}
}

func TestNewSupportsPrometheusExporter(t *testing.T) {
	provider, err := New(config.ObservabilityConfig{
		Enabled:        true,
		Exporter:       "prometheus",
		MetricsPath:    "/metrics",
		ServiceVersion: "test-version",
	}, "pfGoPlus-test", "test", zap.NewNop())
	if err != nil {
		t.Fatalf("expected prometheus exporter to initialize, got %v", err)
	}
	if provider.MetricsHandler() == nil {
		t.Fatal("expected metrics handler for prometheus exporter")
	}
	if err := provider.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown telemetry provider: %v", err)
	}
}
