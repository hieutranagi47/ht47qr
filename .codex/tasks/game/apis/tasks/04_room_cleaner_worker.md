# Task 04: Caro Room Expiration Background Worker

## Objective
Delete expired rooms on a 15-minute schedule through the existing service background lifecycle, using the 24-hour expiry established at creation.

## Architecture and scope
- Follow the repository-root `AGENTS.md` and current `shorten_url/app/metadata.go` and `Service.RunBackground` patterns. These paths and current tooling supersede the standalone layout in `../AGENTS.md`.
- Implement cleanup orchestration in `caro/app/` and SQL in the existing `caro/adapters/{db,postgres}/` repositories. Wire the worker in `caro/module.go`; do not create a separate `room_cleaner` module, executable, cron dependency, or repository.
- Keep `fe/` and embedded `client/api/` assets untouched.

## Implementation
1. Add an application cleanup operation taking a context and injected UTC clock. Delete rooms whose persisted `expires_at <= now`, with cascading player deletion, and return the affected room count (and identifiers if needed for transport eviction).
   - Do not infer expiry solely from status or delete every waiting room. A shorter stale-waiting-room policy is outside this task unless explicitly specified later.
   - Keep the SQLite and PostgreSQL cutoff, cascade, and concurrency semantics consistent. Use context-aware I/O and transactional deletion where multiple statements are needed.
   - Make repeated execution safe and ensure cleanup racing with a join cannot admit a player into an expired/deleted room.
2. Add a cancellable worker loop with a 15-minute ticker. Inject scheduling/clock dependencies as needed to test behavior without real 15-minute sleeps; stop the ticker on exit.
   - Run an initial cleanup pass when background processing starts, then on ticks.
   - Check cancellation before starting work and pass the worker context through cleanup and any notifications.
   - Log transient cleanup failures and retry on the next tick without a busy loop. Treat normal context cancellation as a clean exit; do not continue database work after shutdown.
3. Implement `module.BackgroundModule` on `caro.Module` through `RunBackground(ctx)`. Compose cleanup with any Task 03 transport lifecycle work in this one method and wait for all owned goroutines.
   - `Service.Run` already invokes `RunBackground`; do not start workers in `Init`, `RegisterHttp`, or request handlers.
   - Document that callers embedding `HTTPHandler()` must also run `Service.RunBackground(ctx)`.
   - Preserve caller ownership of database pools and the service's existing graceful shutdown behavior.
4. Coordinate with the signaling transport through an injected interface, keeping `caro/app` independent of `caro/api/socketio`.
   - Evict/close connections for expired rooms, including reconciliation after notification failures. Database deletion must not leave a live connection authorized to relay indefinitely.
   - Read, join, and relay paths enforce expiry independently of the sweep schedule; the cleanup tick is not the source of authorization validity.
5. Use `common/log.FromContext` for structured logs with cleanup duration, deleted-room count, and internal failure details. Do not log credentials or signaling payloads, or introduce a new metrics stack solely for this task.

## Dependencies and acceptance
- Depends on Task 01; integrate live connection eviction with Task 03 when available. Keep the cleanup application service testable without a running transport.
- Test initial execution, periodic scheduling, expiry just before/at/after the cutoff, unexpired waiting-room preservation, cascading deletion, repeatability, database failures, connection eviction/reconciliation, and prompt cancellation with no leaked goroutines.
- Exercise repository semantics against temporary SQLite and a disposable PostgreSQL database, skipping PostgreSQL tests when no test URL is configured. Add a service-level test proving the worker is owned by `RunBackground` and stops on cancellation.
- Regenerate SQL outputs only if queries change, format changed Go files with `go tool gofumpt -w <files>`, run focused tests, then `go test ./...`, relevant `go vet ./...`, and `go test -race ./...` when supported.
