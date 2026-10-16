DROP INDEX caro_room_creation_key;
ALTER TABLE caro_rooms DROP COLUMN recovery_nonce;
ALTER TABLE caro_rooms DROP COLUMN request_fingerprint;
ALTER TABLE caro_rooms DROP COLUMN idempotency_key_hash;
DROP INDEX caro_players_ip_room;
ALTER TABLE caro_room_players DROP COLUMN participant_ip;
