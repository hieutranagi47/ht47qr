# Task 02: Caro Room Application Use Cases and REST APIs

## Objective
Expose room creation and discovery through the existing Echo v5 ordinary HTTP router, backed by the Caro domain and repositories from Task 01.

## Architecture and scope
- Follow the repository-root `AGENTS.md` and current `shorten_url/` patterns. These paths and the current toolchain supersede the legacy standalone layout in `../AGENTS.md`.
- Extend `caro/`; do not create `internal/modules/lobby`, a new executable, or a separate HTTP listener.
- Put rules in `caro/domain/`, orchestration and consumer-owned interfaces in `caro/app/`, and thin request/response mapping in `caro/api/http/`. Inject the repository, clock, and code generator from `caro/module.go`.
- Preserve the P2P scope: the backend manages rooms and signaling, not moves, scores, or win/loss calculation. Leave `fe/` and `client/api/` assets untouched.

## API behavior
1. `POST /api/v1/rooms`
   - Accept `{ displayName, roomType, maxGames }` and validate via domain rules: required name up to 50 characters, `PUBLIC` or `PRIVATE`, and `maxGames` in `{1, 3, 5, 7}`.
   - Atomically create a `WAITING` room with a unique six-character code, expiry at creation plus 24 hours, and a reserved `PLAYER_1` host slot.
   - Return `201` with the room representation and the host reservation information needed by Task 03. Define that private host credential explicitly in the contract before implementation; possession of the public room code alone must not allow claiming an existing host slot.
   - Room creation does not require a socket ID. Task 03 binds the host reservation on connection without creating another player slot.
2. `GET /api/v1/rooms`
   - Return `200` with only unexpired `PUBLIC`, `WAITING` rooms with exactly one occupied/reserved player slot. Define deterministic ordering and an explicit list response schema; return an empty list when none match.
   - Apply visibility and occupancy filtering in the repository query. A private room is discoverable only through its code, not the public list.
3. `GET /api/v1/rooms/{code}`
   - Validate/normalize the code and return `200` with public room details, occupancy, and whether the room is full/joinable.
   - Treat expired rooms as unavailable regardless of whether cleanup has run. Document `400` for malformed input, `404` for missing/expired rooms, and relevant `409` conflicts for mutating operations.
   - Do not expose socket IDs, host credentials, or internal database diagnostics in room/list responses.

## Implementation and integration
- Add `caro/api/http/openapi.yaml`, `oapi-codegen.yaml`, and `generate.go`, following the current strict Echo server generation pattern and pinned generator. Generate `openapi.gen.go`; implement its strict interface in `handler.go`.
- Register handlers through `caro.Module.RegisterHttp(ctx, common.EchoRouter)`. Initialize the application service and handler in the module constructor/lifecycle; preserve initialization and contract registration before transport registration.
- Map domain/application errors through `common` error helpers and the existing error handler. Propagate request contexts to all I/O.
- Reuse `common/http` middleware. The current CORS policy permits same-origin requests: if WASM deployment requires other origins, add an explicit configurable allowlist and preflight tests in the shared configuration/wiring. Do not install a second permissive CORS layer in the feature handler.
- No cross-module contract is needed between room HTTP and signaling inside `caro`. If another feature later consumes Caro, publish the capability via `caro/api/module` and `caro/api/module/client`, register it in `common/module/contracts`, and inject the client interface; do not expose application or repository types as a cross-module API.

## Dependencies and acceptance
- Depends on Task 01. Coordinate the private host reservation contract with Task 03 and document the exact JSON fields and reconnect policy before implementing both transports.
- Test domain rules and application orchestration with injected dependencies, then handler contracts using `httptest`, including malformed bodies, all allowed game counts, collision exhaustion, visibility, occupancy, expiry, error mapping, and credential redaction.
- Add service component coverage through `Service.HTTPHandler()` proving the routes coexist with existing QR, short URL, health, and client routes for both supported database backends. Use disposable PostgreSQL test infrastructure and skip when unavailable.
- Run targeted HTTP generation, format changed Go files with `go tool gofumpt -w <files>`, run focused tests, then `go test ./...` and relevant `go vet ./...`. Do not regenerate unrelated transports or rebuild frontend assets.
