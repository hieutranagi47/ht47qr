# Server-Sent Events

The template runs separate HTTP and HTTPS listeners for SSE. Configure
`SERVER_SSE_PORT` (8886) and `SERVER_SSE_PORT_TLS` (8446), alongside the ordinary
HTTP ports. Both TLS listeners use the PEM content in `TLS_CERT` and `TLS_KEY`.
The Compose template exposes all four ports.

Try the heartbeat demonstration:

```sh
curl -N http://localhost:8886/sse/qrcode/heartbeat
```

It immediately emits a `connected` comment, followed by a heartbeat comment
every 15 seconds. Comments keep the connection alive without dispatching a
browser message. Use the HTTPS listener for browser pages served over HTTPS.

Each module implements `RegisterSSE` and registers paths relative to `/sse`.
Use an empty implementation for modules without streaming routes. SSE routes
are served only on the SSE listeners; `/health` is available on every listener.

Authenticate and authorize the resource, then subscribe before calling
`common/sse.Stream(c, events)`. Defer subscription cleanup. The helper sends
and flushes each `sse.Event`, exits on disconnect/shutdown or channel closure,
and limits each write to 10 seconds to avoid waiting indefinitely on slow
clients. SSE requests skip the normal request timeout and response-body
capture. Ordinary API requests retain their 10-second timeout. Set any proxy
idle timeout above the heartbeat interval and disable response buffering or
compression for streaming routes; the helper sets `X-Accel-Buffering: no`.

The template allows cross-origin requests with `*`, matching htqrcodeagent. Restrict
the CORS origins for your deployment; credentialed browser connections require
explicit allowed origins and credentials support.

Named events require `EventSource.addEventListener(eventName, handler)`;
unnamed events use `onmessage`. Set an immutable event ID if implementing replay
and handle the browser's `Last-Event-ID` in the module before opening the stream.
The helper does not persist events, replay missed messages, or distribute events
across replicas. A shared backend consumer group does not broadcast to every
replica's local subscribers. Choose a per-replica fan-out strategy or a durable
status/replay view. Keep subscription queues bounded and define what happens
when a client falls behind; do not block the backend consumer on a browser.
