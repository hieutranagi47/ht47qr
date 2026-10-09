CREATE SCHEMA IF NOT EXISTS shorten_url;

CREATE TABLE shorten_url.links (
  id TEXT PRIMARY KEY NOT NULL,
  user_id TEXT NOT NULL,
  long_url TEXT NOT NULL,
  short_code TEXT NOT NULL CHECK (length(short_code) = 8),
  idempotency_key TEXT NOT NULL,
  created_at BIGINT NOT NULL,
  expires_at BIGINT NOT NULL,
  deactivated_at BIGINT
);
CREATE UNIQUE INDEX links_short_code_uq ON shorten_url.links (short_code);
CREATE UNIQUE INDEX links_user_idempotency_uq ON shorten_url.links (user_id, idempotency_key);
CREATE INDEX links_user_live_idx ON shorten_url.links (user_id, expires_at) WHERE deactivated_at IS NULL;
CREATE INDEX links_deactivated_idx ON shorten_url.links (deactivated_at) WHERE deactivated_at IS NOT NULL;
CREATE INDEX links_expiry_idx ON shorten_url.links (expires_at);
