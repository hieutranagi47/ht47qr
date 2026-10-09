-- name: GetLinkByUserAndIdempotency :one
SELECT id, user_id, long_url, short_code, idempotency_key, created_at, expires_at, deactivated_at
FROM shorten_url.links
WHERE user_id = $1 AND idempotency_key = $2;

-- name: CountLiveLinksByUser :one
SELECT count(*)
FROM shorten_url.links
WHERE user_id = $1 AND deactivated_at IS NULL AND expires_at > $2;

-- name: InsertLink :one
INSERT INTO shorten_url.links (
  id, user_id, long_url, short_code, idempotency_key, created_at, expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT DO NOTHING
RETURNING id, user_id, long_url, short_code, idempotency_key, created_at, expires_at, deactivated_at;

-- name: GetActiveLinkByCode :one
SELECT id, user_id, long_url, short_code, idempotency_key, created_at, expires_at, deactivated_at
FROM shorten_url.links
WHERE short_code = $1 AND deactivated_at IS NULL AND expires_at > $2;

-- name: DeleteOldLinks :exec
DELETE FROM shorten_url.links
WHERE (deactivated_at IS NOT NULL AND deactivated_at < sqlc.arg(before))
   OR expires_at < sqlc.arg(before);
