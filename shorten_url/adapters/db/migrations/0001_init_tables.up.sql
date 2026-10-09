CREATE TABLE shorten_url_links (
  id TEXT PRIMARY KEY NOT NULL,
  user_id TEXT NOT NULL,
  long_url TEXT NOT NULL,
  short_code TEXT NOT NULL CHECK (length(short_code) = 8),
  idempotency_key TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  expires_at INTEGER NOT NULL,
  deactivated_at INTEGER
);
CREATE UNIQUE INDEX shorten_url_links_short_code_uq ON shorten_url_links (short_code);
CREATE UNIQUE INDEX shorten_url_links_user_idempotency_uq ON shorten_url_links (user_id, idempotency_key);
CREATE INDEX shorten_url_links_user_live_idx ON shorten_url_links (user_id, expires_at) WHERE deactivated_at IS NULL;
CREATE INDEX shorten_url_links_deactivated_idx ON shorten_url_links (deactivated_at) WHERE deactivated_at IS NOT NULL;
CREATE INDEX shorten_url_links_expiry_idx ON shorten_url_links (expires_at);
