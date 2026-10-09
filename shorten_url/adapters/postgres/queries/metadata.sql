-- name: EnqueueMetadata :exec
INSERT INTO shorten_url.metadata (link_id) VALUES ($1) ON CONFLICT DO NOTHING;

-- name: GetMetadata :one
SELECT title, description, image_url, fetched_at
FROM shorten_url.metadata WHERE link_id = $1 AND status = 'ready';

-- name: ClaimMetadata :one
UPDATE shorten_url.metadata
SET attempts = attempts + 1, next_attempt_at = sqlc.arg(lease_until)
WHERE link_id = (
  SELECT m.link_id FROM shorten_url.metadata m
  JOIN shorten_url.links l ON l.id = m.link_id
  WHERE m.status = 'pending' AND m.next_attempt_at <= sqlc.arg(now)
    AND m.attempts < 3 AND l.deactivated_at IS NULL AND l.expires_at > sqlc.arg(now)
  ORDER BY m.next_attempt_at, l.created_at LIMIT 1
  FOR UPDATE OF m SKIP LOCKED
)
RETURNING link_id, attempts;

-- name: GetMetadataJobURL :one
SELECT long_url FROM shorten_url.links WHERE id = $1;

-- name: CompleteMetadata :exec
UPDATE shorten_url.metadata
SET title = $1, description = $2, image_url = $3, fetched_at = $4, status = 'ready'
WHERE link_id = $5 AND status = 'pending' AND attempts = $6;

-- name: FailMetadata :exec
UPDATE shorten_url.metadata
SET status = CASE WHEN attempts >= 3 THEN 'failed' ELSE 'pending' END,
    next_attempt_at = $1
WHERE link_id = $2 AND status = 'pending' AND attempts = $3;

-- name: ExpireMetadataAttempts :exec
UPDATE shorten_url.metadata SET status = 'failed'
WHERE status = 'pending' AND attempts >= 3 AND next_attempt_at <= $1;
