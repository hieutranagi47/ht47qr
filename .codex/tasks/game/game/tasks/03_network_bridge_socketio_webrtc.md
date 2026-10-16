# Task 03: JS Interop Bridge for Socket.IO & WebRTC P2P

## Objective
Implement syscall/js wrappers in Go to communicate with browser Socket.IO signaling and WebRTC DataChannel APIs.

## Instructions
1. Create pkg/network/socket_bridge.go:
   - Bind Go callbacks to JS Socket.IO events (join_room, peer_joined, signal_offer, signal_answer, signal_ice).
2. Create pkg/network/p2p_bridge.go:
   - Bind Go callbacks to WebRTC RTCDataChannel.onmessage.
   - SendP2PMessage(msg P2PMessage): Serialize struct to JSON string and send via dataChannel.
3. Wire network events into gameplay_scene.go:
   - On local move -> call SendP2PMessage(MOVE).
   - On receiving MOVE -> update opponent tile on board.
   - On win -> call SendP2PMessage(ROUND_WIN).
