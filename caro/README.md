# Caro backend

Caro owns rooms and WebRTC signaling. Gameplay, moves, scores, and wins are P2P.

## REST

- `POST /api/v1/rooms`: `{ "displayName": "Alice", "roomType": "PUBLIC", "maxGames": 3 }`.
  Returns `201 { "room": {...}, "hostReservation": { "playerId": "...", "role": "PLAYER_1", "hostToken": "..." } }`.
  The 256-bit host token is reproducible only with the private creation key, hashed at rest, and omitted from public details and request logs. Keep both credentials private.
- `GET /api/v1/rooms`: `{ "rooms": [...] }`, oldest creation time first, ID as tie breaker.
  Only unexpired PUBLIC WAITING rooms with exactly one reserved/occupied slot appear.
- `GET /api/v1/rooms/{code}`: room details, players' display names/roles/connected flags,
  occupancy including reservations, full and joinable. No socket IDs or credentials.

Codes use `ABCDEFGHJKLMNPQRSTUVWXYZ23456789`, exactly six characters, trimmed and
uppercased on input. Names are trimmed, nonblank, at most 50 Unicode characters.
Allowed maxGames: 1, 3, 5, 7. Room lifetime: exactly 24 hours in UTC.
Malformed inputs return 400, absent/expired lookups 404, mutation conflicts 409;
unexpected errors use the shared safe error envelope.

## Per-IP room limits

Each normalized IP may create or participate in at most **50 unexpired rooms total
across WAITING and PLAYING**. Creation reserves the host's waiting-room
quota before a socket connects. Guest joins and host bindings record their
server-resolved IP; every involved IP shares the same combined quota.
Starting an existing waiting room does not consume an additional slot. A room
counts once per IP, even if that IP owns and participates in it.
FINISHED and expired rooms consume no quota. Guest departure releases that
membership; an unbound waiting host retains its reservation.

The IP comes from Echo's configured extractor, never a JSON field. The default
uses the direct network peer and ignores untrusted `X-Forwarded-For` / `X-Real-IP`
headers. IPv4-mapped IPv6 shares the IPv4 quota; canonical IPv6 forms share one
quota. Ports do not affect identity. Clients sharing one resolved IP share the
quota. IPs remain private. Creator ownership remains recorded even if the host
later connects from another IP, which also counts while it participates.

Quota violations use `waiting_room_limit_reached` or `active_room_limit_reached`
(HTTP 409 or the Socket.IO error/ack contract). Membership and status writes roll
back together. SQLite writer transactions and PostgreSQL SERIALIZABLE transactions
prevent races across room creation, guest joins to different rooms, and starts.

Migrations 000002–000003 preserve older rooms and backfill known host IPs. Unknown
legacy IPs cannot be recovered; newly bound hosts and newly joined guests always
record their resolved IP.

## Retrying creation after an interrupted connection

`POST /api/v1/rooms` requires a private UUIDv4 `Idempotency-Key` header. Generate it
before the first request and retain it with the request inputs until the response
is received. Retry using the same key, body and resolved IP. A retry returns 201
with the **same original room representation and host reservation/token**, even
if the room has since started or finished. This is an immutable creation receipt;
use GET for current room state. Retries bypass room quotas and code generation.
Concurrent copies create exactly one room and one host reservation.

Changed inputs with the same key return 409 `idempotency_conflict`. Keys are scoped
to the normalized creator IP and guaranteed for the room's 24-hour lifetime.
A retained expired room returns 410 `idempotency_expired`; cleanup removes its
retry metadata with the room. After expiry, use a fresh key. An IP change creates
a different identity and retry scope.

Treat the UUIDv4 key as a **private recovery credential**: possession from the
same IP recovers the host token. Never publish it or put it in a URL. Only its
SHA-256 hash, the normalized request fingerprint and a random nonce are stored.
The host token is HMAC-SHA256 of that nonce keyed by the private UUIDv4 key, and
only its hash is stored with the player. Retries work after process restarts
without storing plaintext recovery keys/tokens or requiring a process-local cache.
Creation responses use `Cache-Control: no-store`; request/response bodies and
signaling payloads remain excluded from shared logs.

```js
const key = crypto.randomUUID(); // retain this before the first network attempt
const response = await fetch('/api/v1/rooms', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json', 'Idempotency-Key': key },
  body: JSON.stringify({ displayName: 'Alice', roomType: 'PUBLIC', maxGames: 3 })
});
// If interrupted, repeat with the same key and body.
```

## Socket.IO contract

Endpoint: `/api/v1/caro/socket.io/`, same origin as REST. The Go server is
`github.com/zishang520/socket.io/servers/socket/v3 v3.0.6`, which supports browser
`socket.io-client` 4.x (Socket.IO protocol revision 5, Engine.IO 4), including
HTTP long-polling and WebSocket upgrade. The WASM bridge calls the browser JS
client, using `io({ path: '/api/v1/caro/socket.io/', transports: ['websocket'] })`.
The browser connects directly over WebSocket so its handshake, join and heartbeat
remain on one connection. Polling followed by an upgrade requires sticky routing
across requests, which a deployment proxy may not provide. The server retains
polling support for other clients. This transport choice does not provide routing
between different players on different processes; the single-process requirement
below still applies.
Sources: [v3 upgrade guide](https://github.com/zishang520/socket.io/blob/main/docs/UPGRADE.md),
[Socket.IO protocol](https://github.com/socketio/socket.io-protocol),
[Engine.IO protocol](https://github.com/socketio/engine.io-protocol).

Each client event has one object payload and an optional Socket.IO acknowledgement
callback. Success: `{ "ok": true }`; failure: `{ "ok": false, "error": { "message": "...", "slug": "...", "details": [] } }`.
Failures also emit `room_error` with that safe error object. Payloads are bounded
at 64 KiB; binary events are unsupported. Do not send credentials in URLs.

- `join_room { code, displayName?, hostToken? }`: host supplies hostToken; guest
  supplies displayName and omits hostToken. A room code alone cannot bind the
  host reservation. One connection may belong to one room. Duplicate joins to
  that room are idempotent. Server emits `room_joined { code, role }` once.
  When both sockets are connected, the transaction marks PLAYING, then emits
  `peer_joined { code }` to the host (the host initiates WebRTC).
- `signal_offer { code, data: { type: 'offer', sdp: '...' } }`
- `signal_answer { code, data: { type: 'answer', sdp: '...' } }`
- `signal_ice { code, data: { candidate: '...', sdpMid?: string, sdpMLineIndex?: number } }`
  Relays the same `{code,data}` only to the other connected player. Membership,
  PLAYING state, and expiry are checked against persistence for each event.
  Sender socket IDs are never accepted from clients.
- `leave_room { code }` and transport disconnect share idempotent departure.
  The remaining peer receives `peer_disconnected { code }` after commit.

While WAITING, a disconnected host is unbound but retains its reservation and
may reconnect with its token once the old socket disconnects. A guest departure
removes its slot. Guests have no reconnect credential. Once PLAYING, either
player's departure makes the room FINISHED; it never reopens. Finished rooms
remain readable until expiry but cannot join or relay. Clients create a new room
after a terminal disconnect. A delayed old disconnect cannot remove a new socket.

## Lifecycle and deployment

Run **one signaling process per database**. The socket registry and event ordering
are process-local; a shared PostgreSQL database alone does not route notifications.
Multiple replicas require shared relay/routing and coordinated recovery before use.
On background startup, rooms with stale connected sockets are finished and their
socket bindings cleared. Untouched host reservations survive. Socket admission is
unavailable until this recovery completes. Do not start a second instance against
live rooms.

During cold startup, the browser retries the initial signaling handshake for up
to 30 seconds, with a 10-second timeout per attempt and a capped backoff. Startup
503 responses are not cacheable. Once a connection succeeds, disconnects remain
terminal; the bridge does not automatically rejoin an active match. The startup
deadline remains active until `room_joined` confirms admission, even when the
WebSocket handshake and join acknowledgement succeeded. The board displays the
current connection stage and keeps transport errors visible with the information
panel collapsed.

Transient WebRTC `disconnected` states keep signaling and room membership alive.
During peer setup, the original 30-second connection deadline remains in force;
after the data channel opens, a disconnect has 10 seconds to recover. Repeated
state notifications do not extend that window. A recovered connection preserves
the existing data channel and match. Failed/closed peer connections, closed data
channels, signaling loss, and explicit opponent departures remain terminal.
Initial peer failures report that the network may require a TURN relay; a grace
period cannot make a STUN-only deployment work on networks that block direct P2P.

Run the bridge retry/lifecycle tests with `node --test tests/testdata/caro_bridge.test.cjs`.

`Service.Run` owns transport shutdown and cleanup. Embedded users must run
`Service.RunBackground(ctx)` alongside `HTTPHandler()`. Cancellation stops new
admission, closes Socket.IO/Engine.IO connections and waits for owned callbacks;
pools remain caller-owned. The cleaner runs immediately and every 15 minutes,
deletes `expires_at <= now` with cascading players, and reconciles the live registry
even when zero rows were deleted. Reads, joins, and relays enforce expiry without
waiting for cleanup. Failed cleanups are logged and retried on the next tick.

Signaling skips shared body logging and the ten-second HTTP context timeout;
ordinary REST and SSE behavior is preserved. Shared same-origin CORS and rate
limiting still apply. Origin checks also cover WebSocket handshakes. TLS should
terminate at this service or a trusted same-origin proxy.

## Verification

Use `go generate ./caro/adapters/db ./caro/adapters/postgres ./caro/api/http`
with sqlc v1.31.1 and the pinned OpenAPI generator, then `go test ./caro/...`.
`CARO_TEST_POSTGRES_URL` must reference a **disposable database**: Caro tests
serialize using an advisory lock and truncate only `caro.rooms` (cascading players).
The ordinary suite also exercises middleware and service component behavior.

To run the real JS client compatibility test, install `socket.io-client@4.8.1`
in a temporary directory outside this repository, then run:

```sh
CARO_SOCKETIO_CLIENT_DIR=/path/to/temporary/npm-project go test ./tests -run TestCaroBrowserSocketIOClient -count=1
```

Protocol-level WebSocket tests run without Node or npm.

## Browser game

The Ebitengine client lives in the sibling `../caro` Go module. Build with
`cd ../caro && ./scripts/build-wasm.sh` before rebuilding this backend. The
compiled game and matching Go runtime are embedded from `caro/api/wasm/`.

- `GET /caro/boards`: room browser, creation form and join by code.
- `GET /caro/boards/{code}`: host/guest entry and WASM game.
- `GET /caro/wasm/*`: self-hosted client assets, including `caro.wasm`.

The public route uses the six-character room code. Invite URLs contain no host
credentials. The creator's tab retains its host token and creation retry key
in session storage. Room pages use a separate CSP supporting WASM compilation.
The existing `client/api` SPA build is not involved.

X is the host; first to `maxGames/2 + 1` round wins wins the set. Five or more
consecutive stones counts, including blocked ends. Draws replay without scores,
and both players acknowledge the next round. The completed set locks play and
returns to the lobby after five minutes, disconnecting the sockets then. The
existing backend marks a room FINISHED when either player leaves; there is no
server move/score/result endpoint. Production networks that block direct P2P
need deployment-provided TURN configuration (see the client README).

## Docker and Vercel WASM deployment

From the `ht47qr` repository root (with the sibling `../caro` module present):

```sh
task build-wasm
docker compose up --build
# Or run both steps with: task docker-up
```

Compose builds `Dockerfile` and exposes the app at
`http://localhost:8888/caro/boards` (HTTPS at `https://localhost:8443`). It uses
its PostgreSQL service and a persistent named database volume. Runtime `.env`
files are excluded from Docker builds.

### Test TURN locally

`docker compose up --build -d` starts the app, PostgreSQL, and Coturn. Open
`http://localhost:8888/caro/boards` in two separate browser sessions, create a
new room, and invite the second player. Compose defaults to **relay-only** ICE,
so a successful connection exercises TURN even when direct P2P would work.
Existing rooms that already failed cannot be reused.

Firefox blocks loopback STUN/TURN servers by default, including the local
`127.0.0.1:3478` relay. For this same-computer test, open `about:config` in
Firefox, set `media.peerconnection.ice.loopback` to `true`, reload the game,
and create a new room. This is a local development setting; a reachable
non-loopback TURN address avoids the need for it. See
[Mozilla bug 1973521](https://bugzilla.mozilla.org/show_bug.cgi?id=1973521).

TURN listens on `127.0.0.1:3478` over TCP/UDP and publishes UDP relay ports
49160–49200. These follow the [Coturn Docker configuration](https://github.com/coturn/coturn/blob/master/docker/coturn/README.md).
Docker bridge networking works on Docker Desktop without host networking.
Coturn advertises its container relay address: both relay-only browsers reach
the relay endpoints through the same TURN server. Do not advertise `127.0.0.1`
as the external relay address; browsers reject remote loopback ICE candidates.
Inspect it with `docker compose logs --tail=100 turn`.

`CARO_TURN_HOST` defaults to `127.0.0.1`. For two devices on your LAN, set it
to this computer's LAN **IPv4 address**, restart with `docker compose up -d`,
allow port 3478 TCP/UDP and 49160–49200 UDP through the local firewall, and
open the game over trusted HTTPS on both devices. `localhost` HTTP is suitable
only for browsers on the Docker host.

The backend serves `/caro/rtc-config.js` before the game bridge, issuing unique
one-hour HMAC credentials from `CARO_TURN_SECRET`. That shared secret never
appears in browser responses. The Compose default is for local development;
set your own secret for other deployments. Keep relay-only mode for this local
Docker setup. A production relay needs a publicly reachable advertised address
and relay ports before using `CARO_ICE_TRANSPORT_POLICY=all` for mixed direct P2P
and TURN. Outside Compose, supply `CARO_TURN_URLS`
(comma-separated TURN URLs) and `CARO_TURN_SECRET` to enable runtime configuration.

To require evidence of TURN in the existing browser smoke test (Playwright must
be installed outside the repository):

```sh
cd ../caro
CARO_TEST_BASE_URL=http://localhost:8888 CARO_TEST_REQUIRE_RELAY=1 \
  NODE_PATH=/path/to/playwright-project/node_modules node scripts/browser-smoke.cjs
```

This asserts both selected ICE candidates are `relay`, then exercises readiness,
real moves, round wins, and leaving the match.

The WASM task scopes `GOOS=js`, `GOARCH=wasm`, and `CGO_ENABLED=0` to the browser
build. Do not set these globally in Docker or Vercel: the backend is a native
Linux binary built with `CGO_ENABLED=1`. The script copies `wasm_exec.js` from
the same Go compiler that builds `caro.wasm`.

Use `ht47qr` as the Vercel project root. Run `task build-wasm` before uploading
or committing deployment sources, and include both `caro/api/wasm/caro.wasm`
and `caro/api/wasm/wasm_exec.js`. The sibling client module is outside Docker's
build context, so Vercel embeds these prepared assets rather than rebuilding
the client. The Docker build checks that both files are present.

Configure these Vercel runtime settings:

- `PORT=80` (the image default), mapped to `SERVER_PORT` by the image startup command.
- `DATABASE_BACKEND=postgres` (the image default) and `DATABASE_URL` or `POSTGRES_URL`.
- Optionally `DATABASE_URL_UNPOOLED` or `POSTGRES_URL_NON_POOLING` for migrations.
- `TLS_CERT` and `TLS_KEY` containing a matching PEM pair. Both real newlines and
  escaped `\n` are accepted by the image startup command. The bundled development
  pair supports local testing; override it for deployment.

The image supplies the remaining internal listener ports and disables optional
gRPC. Public HTTP, WASM assets, and Socket.IO signaling share the deployment
origin, so no client API URL or WebSocket URL environment variable is needed.
For the ExpressTURN free tier, add these runtime environment variables in the
Vercel project settings. Replace the username and password with your account's
credentials, and use the exact server host/port shown in your dashboard:

```dotenv
CARO_TURN_URLS=turn:free.expressturn.com:3478?transport=udp,turn:free.expressturn.com:3478?transport=tcp
CARO_TURN_USERNAME=sample-turn-username
CARO_TURN_PASSWORD=sample-turn-password
CARO_ICE_TRANSPORT_POLICY=all
```

No secret key is required. ExpressTURN's shared-secret authentication is a
Premium feature; the free tier uses the username and password directly.
Outside Compose, omitting `CARO_TURN_URLS` with provider credentials defaults to
`free.expressturn.com:3478` over UDP and TCP. In Compose, set `CARO_TURN_URLS`
explicitly to replace the local relay and set both credentials in `.env`.

Leave `CARO_TURN_SECRET` unset for hosted credentials. When both username and
password are supplied, they take precedence over the local coturn secret.
`all` allows direct connections and TURN fallback; use `relay` to force TURN
when verifying the deployment. Apply the variables to the intended Vercel
environment and redeploy for the running server to pick them up.

Open `/caro/boards` on the deployed HTTPS origin. The page loads the runtime
configuration from `/caro/rtc-config.js` before starting WebRTC. Credentials
stay out of build assets and the response uses `Cache-Control: no-store`.
Provider credentials are sent to the browser because WebRTC needs them to
connect to TURN; they can be inspected by visitors. With no TURN configuration,
the browser keeps the default public STUN server.
