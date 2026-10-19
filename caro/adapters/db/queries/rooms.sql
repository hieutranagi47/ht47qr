-- name: InsertRoom :execrows
INSERT INTO caro_rooms (id,code,room_type,max_games,status,created_at,expires_at,owner_ip,idempotency_key_hash,request_fingerprint,recovery_nonce) VALUES (sqlc.arg(id),sqlc.arg(code),sqlc.arg(room_type),sqlc.arg(max_games),sqlc.arg(status),sqlc.arg(created_at),sqlc.arg(expires_at),sqlc.narg(owner_ip),sqlc.narg(idempotency_key_hash),sqlc.narg(request_fingerprint),sqlc.narg(recovery_nonce)) ON CONFLICT(code) DO NOTHING;
-- name: RoomByCode :one
SELECT * FROM caro_rooms WHERE code=sqlc.arg(code);
-- name: Players :many
SELECT * FROM caro_room_players WHERE room_id=sqlc.arg(room_id) ORDER BY player_role;
-- name: PublicWaiting :many
SELECT r.* FROM caro_rooms r WHERE room_type='PUBLIC' AND status='WAITING' AND expires_at>sqlc.arg(cutoff) AND (SELECT COUNT(*) FROM caro_room_players p WHERE p.room_id=r.id)=1 ORDER BY created_at,id;
-- name: SaveStatus :exec
UPDATE caro_rooms SET status=sqlc.arg(status) WHERE id=sqlc.arg(id);
-- name: DeletePlayers :exec
DELETE FROM caro_room_players WHERE room_id=sqlc.arg(room_id);
-- name: InsertPlayer :exec
INSERT INTO caro_room_players (id,room_id,display_name,socket_id,player_role,credential_hash,joined_at,participant_ip) VALUES (sqlc.arg(id),sqlc.arg(room_id),sqlc.arg(display_name),sqlc.narg(socket_id),sqlc.arg(player_role),sqlc.arg(credential_hash),sqlc.arg(joined_at),sqlc.narg(participant_ip));
-- name: DeleteExpired :execrows
DELETE FROM caro_rooms WHERE expires_at<=sqlc.arg(cutoff);
-- name: ResetRoomStatus :exec
UPDATE caro_rooms SET status='FINISHED' WHERE EXISTS (SELECT 1 FROM caro_room_players p WHERE p.room_id=caro_rooms.id AND p.socket_id IS NOT NULL);
-- name: ResetSockets :exec
UPDATE caro_room_players SET socket_id=NULL WHERE socket_id IS NOT NULL;
-- name: CountParticipantRooms :one
SELECT count(*) FROM caro_rooms r WHERE r.status IN ('WAITING','PLAYING') AND r.expires_at>sqlc.arg(cutoff) AND r.id<>sqlc.arg(exclude_id) AND (r.owner_ip=sqlc.arg(ip) OR EXISTS (SELECT 1 FROM caro_room_players p WHERE p.room_id=r.id AND p.participant_ip=sqlc.arg(ip)));
-- name: RoomByCreationKey :one
SELECT * FROM caro_rooms WHERE owner_ip=sqlc.arg(owner_ip) AND idempotency_key_hash=sqlc.arg(idempotency_key_hash);
