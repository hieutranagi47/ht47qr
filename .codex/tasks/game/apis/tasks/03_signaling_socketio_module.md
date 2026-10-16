# Task 03: Caro Socket.IO Signaling Transport

## Objective
Add Socket.IO room membership and WebRTC signaling to the existing Caro module. Gameplay remains P2P; the server relays SDP/ICE and manages membership only.

## Architecture and scope
- Follow the repository-root `AGENTS.md` and current module lifecycle. These paths and current Go/tooling requirements supersede the standalone layout in `../AGENTS.md`.
- Add the inbound transport under `caro/api/socketio/`, with membership use cases in `caro/app/` and room invariants in `caro/domain/`. Transport connection IDs and payload DTOs must not become shared domain/database models.
- Use the room repositories from Task 01. Do not create `internal/modules/signaling`, duplicate room persistence, or import another module's adapters.
- Keep Socket.IO connection routing behind a narrow application-consumed notification interface if application orchestration needs to notify peers. Wire its implementation in `caro/module.go` without an import cycle.
- Leave `fe/` and embedded `client/api/` assets untouched.

## Transport integration
1. Choose a Go Socket.IO implementation only after verifying Engine.IO/Socket.IO protocol compatibility with the intended WASM client. Record the compatible versions and endpoint path in the transport documentation; a plain WebSocket server is not equivalent to Socket.IO.
2. Register the endpoint through `caro.Module.RegisterHttp` on the existing ordinary HTTP router. Do not repurpose the separate SSE router or introduce another standalone server/port.
3. Account explicitly for middleware and listener behavior: `common/http/middleware.go` currently has a ten-second request timeout, response logging/buffering, rate limiting, and same-origin CORS; ordinary listeners have a write timeout. Add narrowly scoped Socket.IO handling for polling/upgrades where required, including safe logging and origin validation, while preserving ordinary API protections and existing SSE exceptions.
4. Make connection lifetime follow service cancellation, not merely the handshake request context. Close upgraded connections and wait for transport goroutines during shutdown; HTTP server shutdown alone does not close hijacked sockets. Use the existing `module.BackgroundModule` hook for lifecycle work, composed with Task 04 by one `caro.Module.RunBackground(ctx)` implementation.
5. Document the initial deployment topology. A process-local socket registry requires a single signaling instance or an explicit shared relay/routing design; shared PostgreSQL membership alone cannot deliver events across instances.

## Events and application behavior
- Define and document payloads, acknowledgements/errors, and room-code normalization for all events before implementation. Reuse application behavior and errors rather than embed SQL or business rules in callbacks.
- `join_room`:
  - Reject missing/expired/finished rooms and unauthorized attempts to claim an occupied slot.
  - Bind the reserved host from Task 02 using its private reservation credential; assign a new guest `PLAYER_2`. Define reconnect and credential handling consistently with the REST contract.
  - Atomically enforce at most two players under concurrent joins. Repeated joins by the same connection must not create duplicate membership or duplicate notifications.
  - Join the transport room `room:{code}` only after persistence succeeds; handle transport failure so a failed join does not leave an unintended occupied slot.
  - Emit `room_joined` with the assigned role. Once both players are connected, transition to `PLAYING` and emit `peer_joined` to the host after commit. The host reservation alone is not a connected peer.
- `signal_offer`, `signal_answer`, `signal_ice`:
  - Accept events only from a connection currently admitted to that room and relay only to its other connected peer.
  - Derive sender/membership from server state, not caller-supplied socket IDs. Validate payload shape and size; do not broadcast SDP/ICE globally or log their contents.
- `leave_room` and `disconnect`:
  - Reuse one idempotent leave use case that removes/unbinds membership, updates room state atomically, and emits `peer_disconnected` after commit.
  - Specify host departure, guest departure, reconnection, and terminal-room transitions explicitly before implementation. Do not silently reopen a `PLAYING` room without an agreed rule.
  - A delayed disconnect from an old socket must not remove a newer connection's membership. Reconcile stale membership after process failure according to the documented topology/reconnect policy.
- Expiry applies during joins and relays even if Task 04 has not deleted the rows. Coordinate cleanup with the connection registry so expired rooms stop receiving signaling.

## Dependencies and acceptance
- Depends on Tasks 01 and 02. Resolve host binding, leave/reconnect policy, protocol versions, and deployment assumptions as explicit implementation decisions; they are not already supported by the current server.
- Test application membership with injected persistence/notification dependencies, plus transport integration with a compatible Socket.IO client: host binding, guest join, full-room races, duplicate joins/leaves, cross-room rejection, relay isolation, expiry, reconnect races, and failure compensation.
- Test polling/upgrades through the actual shared middleware and shutdown with open connections. Verify existing REST timeout behavior and SSE behavior remain intact.
- Run focused tests, `go test ./...`, relevant `go vet ./...`, and `go test -race ./...` when supported. Format changed Go files with `go tool gofumpt -w <files>`. Do not add generated REST/gRPC code for Socket.IO events.
