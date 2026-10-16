ALTER TABLE caro_room_players ADD COLUMN participant_ip TEXT;
UPDATE caro_room_players SET participant_ip=(SELECT owner_ip FROM caro_rooms r WHERE r.id=room_id) WHERE player_role='PLAYER_1';
CREATE INDEX caro_players_ip_room ON caro_room_players(participant_ip,room_id) WHERE participant_ip IS NOT NULL;
ALTER TABLE caro_rooms ADD COLUMN idempotency_key_hash TEXT;
ALTER TABLE caro_rooms ADD COLUMN request_fingerprint TEXT;
ALTER TABLE caro_rooms ADD COLUMN recovery_nonce TEXT;
CREATE UNIQUE INDEX caro_room_creation_key ON caro_rooms(owner_ip,idempotency_key_hash) WHERE idempotency_key_hash IS NOT NULL;
