CREATE TABLE shorten_url.metadata (
  link_id TEXT PRIMARY KEY NOT NULL REFERENCES shorten_url.links(id) ON DELETE CASCADE,
  title TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  image_url TEXT NOT NULL DEFAULT '',
  fetched_at BIGINT,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'ready', 'failed')),
  attempts BIGINT NOT NULL DEFAULT 0,
  next_attempt_at BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX metadata_pending_idx ON shorten_url.metadata(next_attempt_at) WHERE status = 'pending';
-- Existing links can also receive previews after upgrading.
INSERT INTO shorten_url.metadata (link_id) SELECT id FROM shorten_url.links;
