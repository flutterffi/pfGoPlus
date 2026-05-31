package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/flutterffi/pfGoPlus/internal/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
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
	metricsHandler http.Handler
}

func New(cfg config.ObservabilityConfig, serviceName, env string, log *zap.Logger) (*Provider, error) {
	if !cfg.Enabled {
		return NewNoop(serviceName), nil
	}

	traceExporter, metricReader, metricsHandler, err := newExporters(cfg.Exporter)
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
		sdkmetric.WithReader(metricReader),
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
		metricsHandler: metricsHandler,
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

func (p *Provider) MetricsHandler() http.Handler {
	return p.metricsHandler
}

func (p *Provider) Shutdown(ctx context.Context) error {
	if err := p.meterProvider.Shutdown(ctx); err != nil {
		return err
	}
	return p.tracerProvider.Shutdown(ctx)
}

func newExporters(name string) (sdktrace.SpanExporter, sdkmetric.Reader, http.Handler, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "stdout":
		traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, nil, nil, fmt.Errorf("create stdout trace exporter: %w", err)
		}
		metricExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
		if err != nil {
			return nil, nil, nil, fmt.Errorf("create stdout metric exporter: %w", err)
		}
		return traceExporter, sdkmetric.NewPeriodicReader(metricExporter), nil, nil
	case "prometheus":
		traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, nil, nil, fmt.Errorf("create stdout trace exporter: %w", err)
		}
		exporter, err := promexporter.New()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("create prometheus exporter: %w", err)
		}
		return traceExporter, exporter, promhttp.Handler(), nil
	default:
		return nil, nil, nil, fmt.Errorf("unsupported telemetry exporter: %s", name)
	}
}
