/* Socket.IO owns signaling only; ordered reliable WebRTC owns gameplay. */
'use strict';
window.caroBridge = {
  connect(notify) {
    const config = window.caroConfig;
    const socket = io({ path: '/api/v1/caro/socket.io/', reconnection: false, autoConnect: false, timeout: 10000 });
    let pc, channel, closed = false, failed = false, joined = false, timeout;
    let connected = false, retry, retryDelay = 500, startupTimeout;
    let pendingICE = [], chain = Promise.resolve();
    const emit = (type, data) => { if (!closed) notify(type, data); };
    const fail = message => {
      if (closed || failed) return;
      failed = true;
      emit('error', message);
      cleanup();
    };
    const signal = (event, data) => {
      socket.timeout(10000).emit(event, { code: config.code, data }, (err, ack) => {
        if (err || !ack?.ok) fail(ack?.error?.message || 'Signaling timed out. Return to rooms and try again.');
      });
    };
    const bindChannel = value => {
      if (channel) { value.close(); return; }
      channel = value;
      channel.onopen = () => { clearTimeout(timeout); emit('open', {}); };
      channel.onmessage = event => {
        if (typeof event.data !== 'string' || event.data.length > 2048) { fail('Invalid peer message'); return; }
        try { emit('message', JSON.parse(event.data)); } catch { fail('Invalid peer message'); }
      };
      channel.onclose = () => fail('Peer connection closed. Create a new room to play again.');
      channel.onerror = () => fail('Peer connection failed. Create a new room to play again.');
    };
    const peer = () => {
      if (pc) return pc;
      pc = new RTCPeerConnection(window.caroRTCConfig || { iceServers: [{ urls: 'stun:stun.l.google.com:19302' }] });
      pc.onicecandidate = event => {
        if (!event.candidate) return;
        const { candidate, sdpMid, sdpMLineIndex } = event.candidate;
        signal('signal_ice', { candidate, ...(sdpMid !== null ? { sdpMid } : {}), ...(sdpMLineIndex !== null ? { sdpMLineIndex } : {}) });
      };
      pc.ondatachannel = event => bindChannel(event.channel);
      pc.onconnectionstatechange = () => {
        if (['failed', 'closed', 'disconnected'].includes(pc.connectionState)) fail('Peer disconnected. Create a new room to play again.');
      };
      timeout = setTimeout(() => fail('Could not connect to the other player. This network may require a TURN relay.'), 30000);
      return pc;
    };
    const serial = fn => payload => {
      if (closed || (payload?.code && payload.code !== config.code)) return;
      chain = chain.then(async () => { if (!closed) await fn(payload); }).catch(() => fail('Could not establish the peer connection. Return to rooms to retry.'));
    };
    const flushICE = async () => {
      for (const candidate of pendingICE) await pc.addIceCandidate(candidate);
      pendingICE = [];
    };
    socket.on('connect', () => {
      connected = true;
      clearTimeout(retry);
      clearTimeout(startupTimeout);
      if (joined) { fail('Signaling connection changed. Return to rooms.'); return; }
      const payload = { code: config.code };
      if (config.hostToken) payload.hostToken = config.hostToken;
      else payload.displayName = config.displayName;
      socket.timeout(10000).emit('join_room', payload, (err, ack) => {
        if (err || !ack?.ok) fail(ack?.error?.message || 'Could not join this room.');
      });
    });
    socket.on('room_joined', data => { joined = true; emit('room_joined', data); });
    socket.on('peer_joined', serial(async () => {
      const connection = peer();
      bindChannel(connection.createDataChannel('caro', { ordered: true }));
      await connection.setLocalDescription(await connection.createOffer());
      signal('signal_offer', { type: connection.localDescription.type, sdp: connection.localDescription.sdp });
    }));
    socket.on('signal_offer', serial(async ({ data }) => {
      const connection = peer();
      await connection.setRemoteDescription(data);
      await flushICE();
      await connection.setLocalDescription(await connection.createAnswer());
      signal('signal_answer', { type: connection.localDescription.type, sdp: connection.localDescription.sdp });
    }));
    socket.on('signal_answer', serial(async ({ data }) => { await peer().setRemoteDescription(data); await flushICE(); }));
    socket.on('signal_ice', serial(async ({ data }) => {
      // ICE can arrive before the SDP offer/answer even over ordered signaling.
      if (!pc?.remoteDescription) pendingICE.push(data);
      else await pc.addIceCandidate(data);
    }));
    socket.on('room_error', error => fail(error.message || 'Room unavailable'));
    // A cold deployment can serve HTTP before signaling recovery has completed.
    // Retry only the initial handshake: reconnecting an admitted player can finish a match.
    socket.on('connect_error', () => {
      if (closed) return;
      if (connected) { fail('Signaling disconnected. Create a new room to play again.'); return; }
      clearTimeout(retry);
      retry = setTimeout(() => { if (!closed) socket.connect(); }, retryDelay);
      retryDelay = Math.min(retryDelay * 2, 2000);
    });
    socket.on('peer_disconnected', () => fail('Opponent left. Create a new room to play again.'));
    socket.on('disconnect', () => { if (!closed) fail('Signaling disconnected. Create a new room to play again.'); });
    function cleanup() {
      if (closed) return;
      closed = true;
      clearTimeout(timeout);
      clearTimeout(retry);
      clearTimeout(startupTimeout);
      window.removeEventListener('pagehide', cleanup);
      if (channel) { channel.onopen = channel.onmessage = channel.onclose = channel.onerror = null; channel.close(); }
      if (pc) { pc.onicecandidate = pc.ondatachannel = pc.onconnectionstatechange = null; pc.close(); }
      socket.removeAllListeners();
      socket.disconnect();
    }
    window.addEventListener('pagehide', cleanup);
    startupTimeout = setTimeout(() => fail('Signaling unavailable. Return to rooms and try again.'), 30000);
    socket.connect();
    return {
      send(data) {
        if (closed || channel?.readyState !== 'open' || channel.bufferedAmount > 65536) return false;
        try { channel.send(data); return true; } catch { return false; }
      },
      close: cleanup
    };
  }
};
