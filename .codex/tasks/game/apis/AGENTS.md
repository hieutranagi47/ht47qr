# AGENTS.md - Backend APIs & Signaling Architecture

Welcome AI Agent. This document defines the architecture, schema definitions, and API contracts for the Backend Service of the Caro P2P Game.

## 1. System Architecture
- **Tech Stack:** Go (1.22+), Modular Monolith, PostgreSQL (`caro` schema), Socket.IO.
- **Role:** Room Management (Lobby), WebRTC Signaling Relay, and 24h Room Expiration Cleanup.
- **Non-Goal:** Backend does NOT compute game moves or determine win/loss logic. All gameplay runs P2P on Ebitengine WASM clients.

## 2. Directory Structure
- `cmd/server/main.go`
- `db/migrations/` (PostgreSQL scripts for `caro` schema)
- `internal/modules/lobby/` (REST APIs for rooms and dashboard)
- `internal/modules/signaling/` (Socket.IO relay for SDP Offer/Answer/ICE)
- `internal/modules/room_cleaner/` (Cron worker for 24h room expiration)

## 3. Database Schema (`caro` Schema)
- `caro.rooms`
  - `id`: UUID / VARCHAR(36) PRIMARY KEY
  - `code`: VARCHAR(10) UNIQUE (6-char room code)
  - `room_type`: 'PUBLIC' | 'PRIVATE'
  - `max_games`: INT (1, 3, 5, 7)
  - `status`: 'WAITING' | 'PLAYING' | 'FINISHED'
  - `created_at`: TIMESTAMPTZ
  - `expires_at`: TIMESTAMPTZ (created_at + 24h)
- `caro.room_players`
  - `id`: UUID / VARCHAR(36) PRIMARY KEY
  - `room_id`: FK to `caro.rooms.id`
  - `display_name`: VARCHAR(50)
  - `socket_id`: VARCHAR(100)
  - `player_role`: 'PLAYER_1' | 'PLAYER_2'

## 4. REST APIs
- `POST /api/v1/rooms`: Create public or private room. Body: `{ displayName, roomType, maxGames }`.
- `GET /api/v1/rooms`: List active public waiting rooms for dashboard.
- `GET /api/v1/rooms/{code}`: Get room info by room code.

## 5. Socket.IO Signaling Protocol
- Client -> Server: `join_room`, `signal_offer`, `signal_answer`, `signal_ice`, `leave_room`.
- Server -> Client: `room_joined`, `peer_joined`, `signal_offer`, `signal_answer`, `signal_ice`, `peer_disconnected`.
