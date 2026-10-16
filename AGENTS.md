# Project instructions

## Scope

- Ignore `fe/`: do not inspect, edit, format, install dependencies, or run frontend commands there unless the user explicitly requests it. Exclude it from repository searches.
- Work from the repository root. Keep changes focused and preserve unrelated local work.
- Do not rebuild frontend assets in `client/api/` for backend changes. These files are embedded in the Go binary and must remain present.

## Project map

This is the `htqrcode` Go module, providing QR generation and URL shortening through an Echo v5 HTTP server, separate SSE listeners, and an optional TLS gRPC listener.

- `cmd/main.go`: executable entry point, environment loading, database selection, and signal handling.
- `service.go`: composition root, module initialization, transport registration, background workers, and graceful shutdown.
- `qrcode/`: message payload rules, QR generation, rendering, and HTTP/gRPC/SSE/module APIs.
- `shorten_url/`: short URL rules, application services, metadata fetching, HTTP previews, and SQLite/PostgreSQL persistence.
- `common/`: shared HTTP, gRPC, SSE, logging, tracing, errors, migrations, and module contracts.
- `client/module.go`: Go-side embedded static asset serving and routing.
- `tests/`: service component tests, including transport behavior and QR decoding.
- `example.env`, `Taskfile.yml`, and the Dockerfiles describe configuration and tooling. `GRPC.md` and `SSE.md` provide transport background; verify details against current handlers, schemas, and tests before reusing examples.

## Architecture and conventions

- Follow the existing `domain/`, `app/`, `adapters/`, and `api/` layout within each feature. Keep business rules in the domain, orchestration and consumer-owned dependency interfaces in the application layer, and database/rendering/network implementations in adapters.
- Keep transport handlers thin. Map request data to application inputs and map results/errors back to the transport. Reuse application behavior across HTTP, gRPC, and module APIs.
- Keep domain models separate from generated transport and database models. Preserve constructor validation and private entity state where established.
- Wire dependencies in module constructors and `service.go`. Use the existing `api/module/client` contracts for cross-module calls and register them through `common/module/contracts`.
- Preserve the module lifecycle: initialize and register contracts before registering transports. Background jobs must stop on context cancellation.
- Propagate contexts through I/O and use the shared logging and error helpers. Preserve public error contracts and keep internal diagnostics out of client responses.
- Keep ordinary HTTP and SSE routers separate. Streaming handlers must flush events and exit on disconnect or shutdown; preserve the existing SSE timeout exceptions.

## Development and validation

- Use the Go version declared in `go.mod` (currently `1.27.1`). SQLite uses `go-sqlite3`, so builds/tests need CGO and a C compiler. Report an unavailable toolchain rather than changing the Go version to work around it.
- Format changed Go files with `go tool gofumpt -w <files>`. `task fmt` formats the entire repository; use it only when that scope is appropriate.
- Run focused tests while iterating, then `go test ./...` for backend code changes. Use `go vet ./...` when relevant, and `go test -race ./...` for concurrency or lifecycle changes when supported.
- Test behavior at the affected layer: domain rules, repository semantics, HTTP contracts, transport parity, or worker shutdown. Existing tests use `testing`, `testify/require`, `httptest`, temporary SQLite databases, and injected dependencies.
- PostgreSQL integration tests require `SHORTEN_URL_TEST_POSTGRES_URL`; otherwise they skip. Set it only to a disposable test database: `TestPostgresShortenURL` truncates short URL tables. Exercise both database adapters when changing persistence semantics.

## Generated code and persistence

- Edit source schemas and regenerate outputs rather than hand-editing generated files:
  - OpenAPI: `openapi.yaml` and `oapi-codegen.yaml` beside each transport's `generate.go`; output is `openapi.gen.go`.
  - Protobuf: `common/api/grpc/errors.proto` and `qrcode/api/grpc/service.proto`; outputs are `*.pb.go`.
  - SQL: `shorten_url/adapters/{db,postgres}/queries/`, `migrations/`, and `sqlc.yaml`; outputs are the corresponding `dbmodels/` packages.
- Run `go generate ./<affected-package>` for targeted generation. `task gen` (also exposed as `task generate`) runs `go generate ./...` and repository-wide formatting.
- Full generation requires `sqlc`, `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` on `PATH`; OpenAPI directives invoke a pinned generator with `go run`. Use the versions pinned in the current directives and Dockerfile.
- Commit relevant regenerated outputs together with their source changes and inspect the diff for unrelated churn.
- Add new numbered migrations for schema evolution. Keep SQLite and PostgreSQL implementations consistent, including idempotency, quotas, expiry, and metadata job claims. Migrations are embedded and applied during module initialization.

## Runtime configuration

- Start locally with `go run ./cmd` after configuring the environment; `cmd` loads `.env` if present. Use `example.env` as the configuration reference and keep credentials out of tracked files.
- `DATABASE_BACKEND` accepts `auto`, `sqlite`, or `postgres`. Automatic selection prefers `DATABASE_URL`, then `POSTGRES_URL`, and otherwise uses SQLite at `SQLITE_PATH` (default `./data/htqrcode.db`). Explicit `postgres` requires a URL; explicit `sqlite` forces local storage.
- Optional direct migration connections use `DATABASE_URL_UNPOOLED` or `POSTGRES_URL_NON_POOLING`. Database pools are owned by the caller supplying them to the service.
- Standalone startup requires all four HTTP/HTTPS/SSE port variables plus `TLS_CERT` and `TLS_KEY` containing actual PEM content with real newlines. An empty `SERVER_GRPC_PORT` disables gRPC.
- Embedded callers using `HTTPHandler()` must also run `RunBackground(ctx)` for durable module jobs; `Service.Run` starts those jobs automatically.
- Docker builds include a frontend stage that reads `fe/`. Use backend Go commands for routine work under the scope above.
