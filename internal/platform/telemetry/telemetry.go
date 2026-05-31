package telemetry

import (
	"context"
	"fmt"
	"strings"

	"github.com/flutterffi/pfGoPlus/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	propagator     propagation.TextMapPropagator
	serviceName    string
}

func New(cfg config.ObservabilityConfig, serviceName, env string, log *zap.Logger) (*Provider, error) {
	if !cfg.Enabled {
		return NewNoop(serviceName), nil
	}

	exporter, err := newExporter(cfg.Exporter)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentName(env),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("build telemetry resource: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagator)

	log.Info("telemetry initialized",
		zap.String("exporter", cfg.Exporter),
		zap.String("service_version", cfg.ServiceVersion),
	)

	return &Provider{
		tracerProvider: tracerProvider,
		propagator:     propagator,
		serviceName:    serviceName,
	}, nil
}

func NewNoop(serviceName string) *Provider {
	tracerProvider := sdktrace.NewTracerProvider()
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	return &Provider{
		tracerProvider: tracerProvider,
		propagator:     propagator,
		serviceName:    serviceName,
	}
}

func (p *Provider) Tracer(name string) trace.Tracer {
	return p.tracerProvider.Tracer(name)
}

func (p *Provider) TracerProvider() trace.TracerProvider {
	return p.tracerProvider
}

func (p *Provider) Propagator() propagation.TextMapPropagator {
	return p.propagator
}

func (p *Provider) Shutdown(ctx context.Context) error {
	return p.tracerProvider.Shutdown(ctx)
}

func newExporter(name string) (sdktrace.SpanExporter, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "stdout":
		exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("create stdout exporter: %w", err)
		}
		return exporter, nil
	default:
		return nil, fmt.Errorf("unsupported telemetry exporter: %s", name)
	}
}
