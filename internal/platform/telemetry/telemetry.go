package telemetry

import (
	"context"
	"fmt"
	"strings"

	"github.com/flutterffi/pfGoPlus/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	propagator     propagation.TextMapPropagator
	serviceName    string
}

func New(cfg config.ObservabilityConfig, serviceName, env string, log *zap.Logger) (*Provider, error) {
	if !cfg.Enabled {
		return NewNoop(serviceName), nil
	}

	traceExporter, metricExporter, err := newExporters(cfg.Exporter)
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
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	otel.SetTextMapPropagator(propagator)

	log.Info("telemetry initialized",
		zap.String("exporter", cfg.Exporter),
		zap.String("service_version", cfg.ServiceVersion),
	)

	return &Provider{
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,
		propagator:     propagator,
		serviceName:    serviceName,
	}, nil
}

func NewNoop(serviceName string) *Provider {
	tracerProvider := sdktrace.NewTracerProvider()
	meterProvider := sdkmetric.NewMeterProvider()
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	return &Provider{
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,
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

func (p *Provider) Meter(name string) metric.Meter {
	return p.meterProvider.Meter(name)
}

func (p *Provider) Propagator() propagation.TextMapPropagator {
	return p.propagator
}

func (p *Provider) Shutdown(ctx context.Context) error {
	if err := p.meterProvider.Shutdown(ctx); err != nil {
		return err
	}
	return p.tracerProvider.Shutdown(ctx)
}

func newExporters(name string) (sdktrace.SpanExporter, sdkmetric.Exporter, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "stdout":
		traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, nil, fmt.Errorf("create stdout trace exporter: %w", err)
		}
		metricExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
		if err != nil {
			return nil, nil, fmt.Errorf("create stdout metric exporter: %w", err)
		}
		return traceExporter, metricExporter, nil
	default:
		return nil, nil, fmt.Errorf("unsupported telemetry exporter: %s", name)
	}
}
