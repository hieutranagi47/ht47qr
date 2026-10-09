CREATE TABLE shorten_url_metadata (
  link_id TEXT PRIMARY KEY NOT NULL REFERENCES shorten_url_links(id) ON DELETE CASCADE,
  title TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  image_url TEXT NOT NULL DEFAULT '',
  fetched_at INTEGER,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'ready', 'failed')),
  attempts INTEGER NOT NULL DEFAULT 0,
  next_attempt_at INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX shorten_url_metadata_pending_idx ON shorten_url_metadata(next_attempt_at) WHERE status = 'pending';
-- Existing links can also receive previews after upgrading.
INSERT INTO shorten_url_metadata (link_id) SELECT id FROM shorten_url_links;
