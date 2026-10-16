DROP INDEX caro.caro_room_creation_key;
ALTER TABLE caro.rooms DROP COLUMN recovery_nonce;
ALTER TABLE caro.rooms DROP COLUMN request_fingerprint;
ALTER TABLE caro.rooms DROP COLUMN idempotency_key_hash;
DROP INDEX caro.caro_players_ip_room;
ALTER TABLE caro.room_players DROP COLUMN participant_ip;
