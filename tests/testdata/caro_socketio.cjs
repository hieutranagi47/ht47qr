// Run by TestCaroBrowserSocketIOClient with CARO_SOCKETIO_CLIENT_DIR pointing at
// a temporary npm project containing socket.io-client@4.8.1.
const assert = require('node:assert/strict');
const { randomUUID } = require('node:crypto');
const { once } = require('node:events');
const { io } = require(process.env.CARO_SOCKETIO_CLIENT_DIR + '/node_modules/socket.io-client');
const base = process.argv[2];
const sockets = [];
const wait = (ms) => new Promise(resolve => setTimeout(resolve, ms));
async function connect() {
  // Respect the shared per-IP request budget while exercising polling handshakes.
  await wait(350);
  const s = io(base, { path: '/api/v1/caro/socket.io/', reconnection: false, timeout: 3000 });
  sockets.push(s);
  await once(s, 'connect');
  if (s.io.engine.transport.name !== 'websocket') await once(s.io.engine, 'upgrade');
  assert.equal(s.io.engine.transport.name, 'websocket');
  return s;
}
function emit(s, event, data) {
  return new Promise((resolve, reject) => s.timeout(3000).emit(event, data, (err, ack) => err ? reject(err) : resolve(ack)));
}
async function create() {
  const response = await fetch(base + '/api/v1/rooms', {
    method: 'POST', headers: { 'content-type': 'application/json', 'Idempotency-Key': randomUUID() },
    body: JSON.stringify({ displayName: 'Alice', roomType: 'PRIVATE', maxGames: 3 })
  });
  assert.equal(response.status, 201);
  return response.json();
}
const watchdog = setTimeout(() => { console.error("Socket.IO client test timed out"); process.exit(1); }, 12000);
(async () => {
  const { room, hostReservation } = await create();
  const host = await connect(), a = await connect(), b = await connect();
  let joined = 0, peers = 0;
  host.on('room_joined', () => joined++);
  host.on('peer_joined', () => peers++);
  assert.equal((await emit(host, 'join_room', { code: room.code.toLowerCase(), hostToken: hostReservation.hostToken })).ok, true);
  const outcomes = await Promise.all([a, b].map(s => emit(s, 'join_room', { code: room.code, displayName: 'Bob' })));
  assert.equal(outcomes.filter(x => x.ok).length, 1);
  const guest = outcomes[0].ok ? a : b, outsider = outcomes[0].ok ? b : a;
  assert.equal((await emit(host, 'join_room', { code: room.code, hostToken: hostReservation.hostToken })).ok, true);
  await wait(20);
  assert.equal(joined, 1); assert.equal(peers, 1);
  const offer = { code: room.code, data: { type: 'offer', sdp: 'browser-offer' } };
  assert.equal((await emit(outsider, 'signal_offer', offer)).ok, false);
  const received = once(guest, 'signal_offer');
  assert.equal((await emit(host, 'signal_offer', offer)).ok, true);
  assert.deepEqual((await received)[0], offer);
  const ice = { code: room.code, data: { candidate: 'candidate:1', sdpMid: '0', sdpMLineIndex: 0 } };
  const iceReceived = once(host, 'signal_ice');
  assert.equal((await emit(guest, 'signal_ice', ice)).ok, true);
  assert.deepEqual((await iceReceived)[0], ice);
  const disconnected = once(host, 'peer_disconnected');
  assert.equal((await emit(guest, 'leave_room', { code: room.code })).ok, true);
  await disconnected;
  assert.equal((await emit(guest, 'leave_room', { code: room.code })).ok, true);
  assert.equal((await emit(host, 'signal_offer', offer)).ok, false);
  const next = await create();
  const original = await connect();
  assert.equal((await emit(original, 'join_room', { code: next.room.code, hostToken: next.hostReservation.hostToken })).ok, true);
  original.disconnect();
  for (let n = 0; n < 30; n++) {
    const state = await (await fetch(base + '/api/v1/rooms/' + next.room.code)).json();
    if (!state.players[0].connected) break;
    await wait(20);
  }
  const replacement = await connect();
  assert.equal((await emit(replacement, 'join_room', { code: next.room.code, hostToken: next.hostReservation.hostToken })).ok, true);
  console.log('Socket.IO 4.8.1: polling, upgrade, membership, relays, leave and reconnect passed');
})().catch(err => { console.error(err); process.exitCode = 1; }).finally(() => {
  clearTimeout(watchdog);
  for (const s of sockets) s.disconnect();
  // Node fetch may keep idle sockets alive after all assertions have passed.
  setTimeout(() => process.exit(process.exitCode || 0), 100);
});
