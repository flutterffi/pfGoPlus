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
- real `buf`-generated protobuf and gRPC Go stubs
- BFF-style HTTP edge adapter over Todo contracts
- OpenTelemetry metrics for HTTP and gRPC request volume, errors, and latency
- optional Prometheus metrics endpoint for runtime scraping

## Quick Start

```bash
make tidy
make proto
make wire
make run
go run ./cmd/grpcserver
```

Switch config profile with layered files:

```bash
PFGO_APP_ENV=local make run
PFGO_APP_ENV=test make run
PFGO_APP_ENV=docker make run
```

Start the full local stack with OTLP collector and Prometheus:

```bash
make compose-up
```

Switch HTTP to gRPC-backed Todo mode:

```bash
PFGO_TODO_BACKEND_MODE=grpc \
PFGO_GRPC_CLIENT_TARGET=127.0.0.1:9090 \
make run
```

Expose Prometheus metrics:

```bash
PFGO_OBSERVABILITY_EXPORTER=prometheus \
PFGO_OBSERVABILITY_METRICS_PATH=/metrics \
make run
```

Export traces and metrics to an OTLP collector:

```bash
PFGO_OBSERVABILITY_EXPORTER=otlp \
PFGO_OBSERVABILITY_OTLP_ENDPOINT=127.0.0.1:4317 \
PFGO_OBSERVABILITY_OTLP_INSECURE=true \
make run
```

Use an explicit config file when needed:

```bash
PFGO_CONFIG_FILE=./configs/config.test.yaml make run
```

Open:

- `GET /health`
- `POST /api/v1/auth/login`
- `GET /api/v1/todos` (requires Bearer token)
- `POST /api/v1/todos` (requires Bearer token)
- `GET /api/v1/users/me` (requires Bearer token)
- `GET /api/v1/users` (admin only)
- `POST /api/v1/users` (admin only)
- `PATCH /api/v1/users/:id` (admin only, update role/status/profile/password)
- `GET /api/v1/audit/logs` (admin only)
  supports `actor_username`, `action`, `resource`, `status`, `trace_id`, `limit`, `offset`
- `GET /api/v1/roles` (admin only)
- gRPC health service on `:9090`
- gRPC `todo.v1.TodoService` with `ListTodos` and `CreateTodo`
- Prometheus UI on `http://127.0.0.1:9091` when using `make compose-up`

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
  configs/                    # base + profile-layered config files
  internal/bootstrap/         # wire-style assembly
  internal/app/               # bootstrap and lifecycle
  internal/config/            # config model and loader
  internal/modules/auth/      # JWT login and auth middleware
  internal/modules/audit/     # admin audit log module
  internal/modules/role/      # role catalog and permissions module
  internal/modules/user/      # user identity and RBAC module
  internal/modules/todo/      # demo business module
  internal/platform/          # database, logger, telemetry
  internal/transport/httpx/   # HTTP transport and middleware
  internal/transport/grpcx/   # gRPC server and interceptors
  api/proto/todo/v1/          # proto contract and generated Go stubs
  docs/architecture.md        # microservice evolution notes
  deployments/               # local observability stack configs
  tools/                      # build tool tracking
  Dockerfile                  # container image build
  docker-compose.yml          # local multi-process stack
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
- buf-based protobuf generation pipeline
- switchable Todo backend: local service or gRPC client
- BFF-style HTTP adapter around Todo API
- OpenTelemetry metrics pipeline
- Prometheus scraping mode via `/metrics`
- real Wire generation workflow via `make wire`
- OTLP exporter support for traces and metrics
- local Docker Compose stack with gRPC, OTEL Collector, and Prometheus
- layered config loading via base, profile, and env overrides
- persisted user module with RBAC-aware auth and admin APIs
- user lifecycle management with update and disable flows
- audit log capture and admin query endpoint
- role-to-permission authorization for admin and member APIs
- audit log filtering and pagination
- persisted role catalog with seeded permissions

Next milestone:

- business protobuf contracts
- grpc-gateway or BFF-style edge transport
- richer service discovery
