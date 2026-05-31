# pfGoPlus

`pfGoPlus` is a production-style Go backend starter focused on the mainstream monolith path first:

- Gin for HTTP APIs
- GORM for persistence
- Viper for configuration
- Zap for structured logging
- unified JSON responses
- global error handling
- trace-aware request logs
- JWT auth for protected business APIs
- Docker and GitHub Actions delivery basics
- OpenTelemetry tracing on HTTP and gRPC
- Wire-style dependency assembly for HTTP and gRPC entrypoints
- working Todo gRPC service scaffold with round-trip tests

## Quick Start

```bash
make tidy
make run
go run ./cmd/grpcserver
```

Open:

- `GET /health`
- `POST /api/v1/auth/login`
- `GET /api/v1/todos` (requires Bearer token)
- `POST /api/v1/todos` (requires Bearer token)
- gRPC health service on `:9090`
- gRPC `todo.v1.TodoService` with `ListTodos` and `CreateTodo`

Example request:

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -H 'X-Trace-ID: local-demo-trace' \
  -d '{"username":"admin","password":"admin123"}'
```

Then call a protected route:

```bash
curl -X POST http://127.0.0.1:8080/api/v1/todos \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <access_token>' \
  -d '{"title":"Ship scaffold","description":"finish the second milestone"}'
```

## Structure

```text
pfGoPlus/
  cmd/server/                 # process entrypoint
  cmd/grpcserver/             # grpc process entrypoint
  configs/                    # viper config files
  internal/bootstrap/         # wire-style assembly
  internal/app/               # bootstrap and lifecycle
  internal/config/            # config model and loader
  internal/modules/auth/      # JWT login and auth middleware
  internal/modules/todo/      # demo business module
  internal/platform/          # database, logger, telemetry
  internal/transport/httpx/   # HTTP transport and middleware
  internal/transport/grpcx/   # gRPC server and interceptors
  api/proto/todo/v1/          # service contract and Go-side grpc scaffolding
  docs/architecture.md        # microservice evolution notes
  Dockerfile                  # container image build
  .github/workflows/go.yml    # CI smoke test
```

## Roadmap

Current milestone:

- standard Gin monolith scaffold
- SQLite + GORM persistence
- request trace ID propagation
- unified success and error responses
- JWT login and protected routes
- Dockerfile and GitHub Actions CI
- OpenTelemetry trace propagation
- gRPC health server scaffold
- Wire-style dependency assembly
- Todo gRPC service implementation and tests

Next milestone:

- business protobuf contracts
- generated protobuf stubs via protoc or buf
- generated Wire bootstrap
- OpenTelemetry metrics exporter
