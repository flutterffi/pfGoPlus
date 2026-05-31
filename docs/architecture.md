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

## Current Progress

The repository now includes the first microservice-ready step:

- `cmd/grpcserver/` for a standalone gRPC process
- `internal/transport/grpcx/` for server setup and interceptors
- `internal/platform/telemetry/` for OpenTelemetry bootstrap
- `internal/bootstrap/` for Wire-style dependency assembly
- `api/proto/todo/v1/` for the Todo service contract
- `internal/modules/todo/grpc_service.go` for business RPC implementation
- request metrics for HTTP and gRPC latency/error visibility

## Folder Strategy

The current folder split leaves room for deeper service extraction later:

- `cmd/` for app entrypoints
- `internal/modules/` for business modules
- `internal/platform/` for infrastructure adapters
- `internal/transport/` for HTTP today and gRPC tomorrow

## Next Step

When a business module is ready to cross process boundaries, the next change should be:

- define a real protobuf contract under `api/proto/`
- generate protobuf and gRPC stubs from the proto file
- move one module behind the new contract
- emit spans and metrics to a real collector instead of stdout
