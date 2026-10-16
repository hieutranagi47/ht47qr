CREATE SCHEMA IF NOT EXISTS caro;
CREATE TABLE caro.rooms (
 id TEXT PRIMARY KEY,
 code TEXT NOT NULL UNIQUE CHECK(code ~ '^[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{6}$'),
 room_type TEXT NOT NULL CHECK(room_type IN ('PUBLIC','PRIVATE')),
 max_games INTEGER NOT NULL CHECK(max_games IN (1,3,5,7)),
 status TEXT NOT NULL CHECK(status IN ('WAITING','PLAYING','FINISHED')),
 created_at TIMESTAMPTZ NOT NULL,
 expires_at TIMESTAMPTZ NOT NULL CHECK(expires_at=created_at+INTERVAL '24 hours')
);
CREATE TABLE caro.room_players (
 id TEXT PRIMARY KEY,
 room_id TEXT NOT NULL REFERENCES caro.rooms(id) ON DELETE CASCADE,
 display_name TEXT NOT NULL CHECK(length(trim(display_name)) BETWEEN 1 AND 50),
 socket_id TEXT UNIQUE CHECK(socket_id IS NULL OR length(socket_id) BETWEEN 1 AND 100),
 player_role TEXT NOT NULL CHECK(player_role IN ('PLAYER_1','PLAYER_2')),
 credential_hash TEXT NOT NULL CHECK((player_role='PLAYER_1' AND length(credential_hash)=64) OR (player_role='PLAYER_2' AND credential_hash='')),
 joined_at TIMESTAMPTZ NOT NULL,
 UNIQUE(room_id,player_role)
);
CREATE INDEX caro_rooms_expiry ON caro.rooms(expires_at);
CREATE INDEX caro_rooms_public_waiting ON caro.rooms(created_at,id) WHERE room_type='PUBLIC' AND status='WAITING';
