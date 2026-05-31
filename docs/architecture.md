# Monolith Now, Microservices Later

## Stage 1

Use a standard monolith with:

- Gin for HTTP routing and middleware
- GORM for persistence
- Viper for config loading
- Zap for structured logs
- Trace ID propagation from ingress to logs and responses

This stage keeps delivery speed high while the domain is still changing.

## Stage 2

When the service boundary becomes stable, evolve with:

- gRPC for service-to-service contracts
- Wire for dependency graph assembly
- OpenTelemetry for tracing and metrics

The current folder split already leaves room for this path:

- `cmd/` for app entrypoints
- `internal/modules/` for business modules
- `internal/platform/` for infrastructure adapters
- `internal/transport/` for HTTP today and gRPC tomorrow
