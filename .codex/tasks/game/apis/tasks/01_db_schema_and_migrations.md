# Task 01: Caro Domain, Persistence, and Module Foundation

## Objective
Prepare room persistence for the existing `htqrcode` service. This is a future implementation task; do not build a separate server or change existing QR/short URL behavior.

## Architecture and scope
- This task follows the repository-root `AGENTS.md` and current code. Its package paths and Go/tooling requirements supersede the legacy standalone layout and Go 1.22 guidance in `../AGENTS.md`.
- Use one top-level `caro/` feature module for room management, signaling, and cleanup. These are capabilities of the same bounded context, not three modules sharing database tables.
- Follow `shorten_url/module.go`: `caro/domain/` owns rules and private entity state; `caro/app/` owns use cases and narrow repository interfaces; `caro/adapters/{db,postgres}/` owns SQL and row mapping; `caro/api/` owns inbound transports.
- Use the Go version in `go.mod`. Keep `fe/` and the embedded assets in `client/api/` untouched.

## Implementation
1. Define room and player entities in `caro/domain/`, with constructors and behavior that enforce:
   - A unique six-character room code; document its alphabet and normalization consistently across transports.
   - Room types `PUBLIC` and `PRIVATE`, statuses `WAITING`, `PLAYING`, and `FINISHED`, and `maxGames` in `{1, 3, 5, 7}`.
   - A required display name of at most 50 characters, two player slots, and roles `PLAYER_1` and `PLAYER_2`.
   - UTC creation/expiry timestamps and a 24-hour lifetime from creation. Expired rooms are unavailable even before physical cleanup.
2. Add numbered migrations under `caro/adapters/postgres/migrations/` and `caro/adapters/db/migrations/`:
   - PostgreSQL uses `caro.rooms` and `caro.room_players`; SQLite uses collision-free table names such as `caro_rooms` and `caro_room_players` with equivalent semantics.
   - Rooms store `id`, `code`, `room_type`, `max_games`, `status`, `created_at`, and `expires_at`.
   - Players store `id`, `room_id`, `display_name`, nullable `socket_id`, `player_role`, and `joined_at`. A reserved host slot can exist before its socket connects.
   - Enforce unique room codes, valid enum/range values, one role per room, unique connected socket membership, and foreign keys with cascading player deletion. Ensure SQLite foreign keys are actually enabled on the connections used.
   - Index expiry and public waiting-room lookup predicates; avoid redundant indexes already covered by uniqueness constraints.
3. Put use-case-oriented persistence interfaces in `caro/app/`, with context-aware operations for atomic room/host creation, active lookup, public waiting-room listing, joining/leaving, and expired-room deletion. Final signatures should follow the use cases rather than expose generic setters.
   - Create the room and host reservation in one transaction. Retry code collisions a bounded number of times in the application layer, using an injected generator.
   - Serialize competing joins so two concurrent requests cannot claim the guest slot or exceed capacity. Membership changes and status transitions must commit together.
   - Map adapter errors to application errors for not found, expired/unavailable, full, and code conflict. Keep SQL errors out of public responses.
4. Add `queries/`, `sqlc.yaml`, and `generate.go` beside each adapter, following the existing database adapters. Generate `dbmodels/` and map rows to domain state; do not share database structs with HTTP or Socket.IO payloads.
5. Add `caro/module.go` implementing `common/module.Module`, a `Caro` module name in `common/module/module.go`, and registration in `service.go` before the client module.
   - Mirror the existing constructors for SQLite and PostgreSQL, PostgreSQL precedence, optional direct migration pool, and caller-owned database lifetimes.
   - Embed migrations and run the shared migration helpers during `Init(ctx)`, using `caro` as the migration namespace.
   - At this stage transport hooks can be no-ops until Tasks 02 and 03. Do not start workers or listeners in `Init`.
   - Preserve startup without a database by omitting the database-dependent module, as the service currently does for URL shortening.

## Dependencies and acceptance
- Task 01 is the foundation for Tasks 02–04. It owns all room persistence; later tasks extend its application operations and queries rather than introduce separate repositories.
- Test domain validation, transactional rollback, code collisions, concurrent capacity enforcement, expiry boundaries, cascades, and repeatable migration initialization with temporary SQLite databases and a disposable PostgreSQL database. Use a clearly named Caro test database variable; skip PostgreSQL tests when unset and never truncate unrelated schemas.
- Follow the pinned generation/tool versions already used by the project. Run targeted `go generate` for the new adapters, format changed Go files with `go tool gofumpt -w <files>`, run focused tests, then `go test ./...` and relevant `go vet ./...`. Run race checks when exercising concurrent membership. Report unavailable tools without changing `go.mod`.
