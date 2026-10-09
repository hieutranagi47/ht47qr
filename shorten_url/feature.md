# Shorten URL

The `shorten_url` module ports the original `tiny_url` feature to this service's
SQLite database. Startup applies its embedded migrations to the database passed
through `ExternalServices.Database`. The caller owns and closes the connection;
`cmd/main.go` supplies the database configured by `SQLITE_PATH`.

## HTTP API

- `POST /shorten-url` accepts `{"long_url":"https://example.com/path"}` and requires
  a nonblank `Idempotency-Key` header. It returns HTTP 201 with `short_url`,
  `long_url`, and `expires_at`. `short_url` is a relative `/r/{short_code}` URL.
- `GET /r/{short_code}` returns HTTP 307 with the original URL in `Location`.
  Missing, expired, or deactivated links return HTTP 404.
- Both endpoints are public. There is no authentication middleware, token
  validation, auth environment configuration, or identity/plan request field.
- Anonymous links expire after 30 days and have no per-user quota. All anonymous
  requests use the nil UUID as their internal owner, so idempotency keys are
  global for anonymous creation. Use a fresh random key for each new creation;
  repeating a key returns the original link, matching the source behavior.
- Errors use the shared `message`, `slug`, and `details` JSON contract.

For example:

```sh
curl -X POST http://localhost:8080/shorten-url \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: 36d85522-61e6-478e-8342-418c154b7d53' \
  -d '{"long_url":"https://example.com/path"}'
```

## Persistence and concurrency

SQLite stores links in `shorten_url_links`, with UUIDs as text and UTC timestamps
as integer Unix nanoseconds. Unique indexes protect short codes and
`(user_id, idempotency_key)`; live-user, expiry, and deactivation indexes support
lookups and cleanup. Queries and database models are generated with `sqlc`.
Creation uses the shared `common.UpdateInSQLiteTx` helper, which acquires the
SQLite writer lock with `BEGIN IMMEDIATE` before reading and retries busy/locked
transactions. This preserves idempotency and quotas across concurrent pools.

`app.Service.Cleanup` removes links deactivated more than 365 days ago, and
links expired more than 365 days ago, matching the source retention behavior.
Cleanup is an explicit operation; this module does not start a background job.

## Future authentication integration

The application retains `UserID`, `Plan`, and the original plan rules for an
internal caller: free has 5 live links and a 30-day lifetime, pro has 100 and
180 days, and ultimate has 3,000 and 365 days. When auth is implemented, pass the
validated identity and purchased plan from the HTTP boundary to `CreateInput`.
The domain and SQLite repository already support those rules. Anonymous calls
always use the free lifetime and skip the per-user quota.

## Regeneration

```sh
go generate ./shorten_url/api/http
go generate ./shorten_url/adapters/db
```

The HTTP contract is `api/http/openapi.yaml`; its strict Echo 5 server is generated
with the same pinned `oapi-codegen` version as the QR module. Database generation
requires `sqlc` on PATH. Run `go test ./shorten_url/...` for the domain and SQLite
component tests, including persistence, concurrent idempotency, quota enforcement,
redirects, expiry, deactivation, cleanup, and safe HTTP errors.
