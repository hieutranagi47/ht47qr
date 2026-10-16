-- Existing rooms have no recorded owner IP; they expire under the original policy.
ALTER TABLE caro_rooms ADD COLUMN owner_ip TEXT;
CREATE INDEX caro_rooms_owner_status ON caro_rooms(owner_ip,status,expires_at) WHERE owner_ip IS NOT NULL;
