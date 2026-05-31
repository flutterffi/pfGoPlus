# pfGoPlus

`pfGoPlus` is a production-style Go backend starter focused on the mainstream monolith path first:

- Gin for HTTP APIs
- GORM for persistence
- Viper for configuration
- Zap for structured logging
- unified JSON responses
- global error handling
- trace-aware request logs

## Quick Start

```bash
make tidy
make run
```

Open:

- `GET /health`
- `GET /api/v1/todos`
- `POST /api/v1/todos`

Example request:

```bash
curl -X POST http://127.0.0.1:8080/api/v1/todos \
  -H 'Content-Type: application/json' \
  -H 'X-Trace-ID: local-demo-trace' \
  -d '{"title":"Ship scaffold","description":"finish the first milestone"}'
```

## Structure

```text
pfGoPlus/
  cmd/server/                 # process entrypoint
  configs/                    # viper config files
  internal/app/               # bootstrap and lifecycle
  internal/config/            # config model and loader
  internal/modules/todo/      # demo business module
  internal/platform/          # database and logger adapters
  internal/transport/httpx/   # response model, router, middleware
  docs/architecture.md        # microservice evolution notes
```

## Roadmap

Current milestone:

- standard Gin monolith scaffold
- SQLite + GORM persistence
- request trace ID propagation
- unified success and error responses

Next milestone:

- gRPC transport
- Wire dependency injection
- OpenTelemetry tracing and metrics
