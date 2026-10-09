# Shorten URL

The `shorten_url` module supports SQLite and PostgreSQL. Startup applies embedded
migrations for the selected backend. `cmd/main.go` selects PostgreSQL when
`DATABASE_URL` or `POSTGRES_URL` is set (`DATABASE_URL` takes precedence);
otherwise it keeps the existing SQLite database configured by `SQLITE_PATH`.
An invalid PostgreSQL configuration fails startup rather than switching storage.

Set these in Vercel's runtime environment settings:

- `DATABASE_URL` (or `POSTGRES_URL`): the pooled Neon PostgreSQL connection URL.
- `DATABASE_URL_UNPOOLED` (or `POSTGRES_URL_NON_POOLING`): optional direct URL
  for startup migrations. Without it, migrations use the application pool.

For embedding, inject `ExternalServices.Database` for SQLite or
`ExternalServices.Postgres` for a pgx pool. PostgreSQL takes precedence when both
are supplied. `ExternalServices.PostgresMigrations` optionally supplies a separate
migration pool. The caller owns and closes all database connections.

Each backend has independent storage. Existing SQLite files are preserved;
this change does not copy existing links into PostgreSQL. Database credentials
must be supplied at runtime rather than embedded in Docker images.

## HTTP API

- `POST /shorten-url` accepts `{"long_url":"https://example.com/path"}` and requires
  a nonblank `Idempotency-Key` header. It returns HTTP 201 with `short_url`,
  `long_url`, and `expires_at`. `short_url` is a relative `/r/{short_code}` URL.
- `GET /r/{short_code}` returns HTTP 307 with the original URL in `Location`
  for browsers, search crawlers, and unknown clients. Known social preview bots
  receive HTTP 200 metadata HTML when a nonempty snapshot is ready; otherwise
  they also receive the redirect.
  Missing, expired, or deactivated links return an HTML page with HTTP 404 and
  the frontend's `/images/f404_light.png` illustration.
- Both endpoints are public. There is no authentication middleware, token
  validation, auth environment configuration, or identity/plan request field.
- Anonymous links expire after 7 days and have no per-user quota. All anonymous
  requests use the nil UUID as their internal owner, so idempotency keys are
  global for anonymous creation. Use a fresh random key for each new creation;
  repeating a key returns the original link, matching the source behavior.
- Errors other than the short-link 404 page use the shared `message`, `slug`, and
  `details` JSON contract.

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

PostgreSQL stores the same link and metadata fields in `shorten_url.links` and
`shorten_url.metadata`, with BIGINT Unix nanosecond timestamps and generated
pgx queries. Migrations follow the `ht47` module schema pattern and release
their dedicated connections without closing the caller's pool. Creation uses
`common.UpdateInReadCommittedTx` and a transaction advisory lock per user (or
per anonymous idempotency key), preserving quotas and idempotency across service
instances. Link creation and metadata enqueue commit together. Metadata claims
use `FOR UPDATE SKIP LOCKED` so concurrent workers claim different pending jobs.

`app.Service.Cleanup` removes links deactivated more than 365 days ago, and
links expired more than 365 days ago, matching the source retention behavior.
Cleanup remains an explicit operation. Metadata enrichment runs separately as
a cancellable background worker.


## Social previews

Creation atomically inserts a pending job in `shorten_url_metadata` alongside
its link, then returns without waiting for the destination website. Replaying
an idempotency key preserves the snapshot and does not enqueue another job.
The new migration also queues existing links; only active links are fetched.

`Service.Run` starts the metadata worker automatically and cancels it during
shutdown. Applications embedding `Service.HTTPHandler()` must also run
`Service.RunBackground(ctx)` with their lifecycle context. Jobs persist across
restarts. The worker polls once per second when idle, claims one job at a time
with a 30-second lease, and allows three attempts with minute-based backoff.
Expired leases can be reclaimed; attempt checks prevent stale completions.

The fetcher prefers Open Graph title/description/image, then Twitter equivalents,
with HTML title and meta description as fallbacks. Relative image references
are resolved against the final page URL and its HTML base URL. Only the absolute
HTTP(S) **image URL** is stored; no image is downloaded. The stored snapshot has a
fetch timestamp and is not automatically refreshed.

Outbound fetches have an eight-second HTTP timeout, a five-redirect limit, and a
1 MiB HTML limit. Requests reject credentials, nonstandard ports, and non-public
IP destinations. DNS results are checked and the validated IP is dialed directly
for every connection, including redirects. Environment proxies are disabled.
HTML is parsed without executing JavaScript, so pages that only publish metadata
through client-side rendering may have no usable snapshot.

The HTTP adapter matches an explicit, case-insensitive User-Agent allowlist:
Facebook (`facebookexternalhit`, `Facebot`), X (`Twitterbot`), LinkedIn,
Slack link expansion, Discord, Telegram, WhatsApp, Skype previews and Pinterest.
Search crawler identifiers take precedence and receive the redirect. This is a
presentation heuristic, not verified bot identity or an authorization mechanism.

Preview HTML uses `html/template` escaping, an empty body, Open Graph and Twitter
card tags, a canonical/OG URL pointing to the destination, and `noindex`. It is
intended for social sharing, not as a separately indexed search result. Missing,
empty, failed, or unreadable snapshots fall back to the redirect. All resolve
responses use `Vary: User-Agent` and `Cache-Control: no-store`; link activity is
checked before considering metadata, preserving the existing 404 behavior.

## Future authentication integration

The application retains `UserID`, `Plan`, and the original plan rules for an
internal caller: free has 5 live links and a 30-day lifetime, pro has 100 and
180 days, and ultimate has 3,000 and 365 days. When auth is implemented, pass the
validated identity and purchased plan from the HTTP boundary to `CreateInput`.
The domain and both database adapters support those rules. Anonymous calls
always use a 7-day lifetime and skip the per-user quota.

## Regeneration

```sh
go generate ./shorten_url/api/http
go generate ./shorten_url/adapters/db
go generate ./shorten_url/adapters/postgres
```

The HTTP contract is `api/http/openapi.yaml`; its strict Echo 5 server is generated
with the same pinned `oapi-codegen` version as the QR module. Database generation
requires `sqlc` on PATH. Run `go test ./shorten_url/...` for the domain and SQLite
component tests, including persistence, concurrent idempotency, quota enforcement,
redirects, expiry, deactivation, cleanup, and safe HTTP errors.


Run the PostgreSQL integration test against a disposable database (it truncates
the module tables):

```sh
SHORTEN_URL_TEST_POSTGRES_URL='postgres://postgres:test-only@localhost:5432/qr_code_test?sslmode=disable' \
  go test -race ./shorten_url/...
```

This additionally verifies PostgreSQL migrations, restart persistence, concurrent
creation across pools, quotas, code collisions, preview jobs, expiry, and cleanup.
