const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const { join } = require('node:path');
const { test } = require('node:test');
const vm = require('node:vm');

const source = readFileSync(join(__dirname, '../../caro/api/wasm/bridge.js'), 'utf8');

function setup() {
  let now = 0, nextID = 0, options, connects = 0, disconnects = 0;
  const timers = new Map(), handlers = new Map(), pageEvents = new Map();
  const notifications = [], statuses = [], sent = [], warnings = [];
  const channels = [];
  class PeerConnection {
    createDataChannel() {
      const channel = { readyState: 'connecting', bufferedAmount: 0, close() {} };
      channels.push(channel);
      return channel;
    }
    async createOffer() { return { type: 'offer', sdp: 'test-offer' }; }
    async setLocalDescription(value) { this.localDescription = value; }
    close() {}
  }
  const socket = {
    io: { engine: { transport: { name: 'websocket' } } },
    on(event, fn) { handlers.set(event, fn); },
    connect() { connects++; },
    disconnect() { disconnects++; },
    removeAllListeners() { handlers.clear(); },
    timeout() { return this; },
    emit(event, payload, ack) { sent.push({ event, payload }); ack(null, { ok: true }); }
  };
  const window = {
    caroConfig: { code: 'ABC234', hostToken: 'private-host-token' },
    addEventListener(event, fn) { pageEvents.set(event, fn); },
    removeEventListener(event) { pageEvents.delete(event); }
  };
  vm.runInNewContext(source, {
    window,
    RTCPeerConnection: PeerConnection,
    console: { warn(...args) { warnings.push(args); } },
    io(opts) { options = opts; return socket; },
    setTimeout(fn, delay) { const id = ++nextID; timers.set(id, { fn, at: now + delay }); return id; },
    clearTimeout(id) { timers.delete(id); }
  });
  const bridge = window.caroBridge.connect((type, data) => type === 'status' ? statuses.push(data) : notifications.push({ type, data }));
  return {
    bridge, notifications, statuses, sent, warnings, options, channels,
    get connects() { return connects; },
    get disconnects() { return disconnects; },
    get pendingTimers() { return timers.size; },
    fire(event, data) { handlers.get(event)?.(data); },
    hide() { pageEvents.get('pagehide')?.(); },
    tick(ms) {
      const end = now + ms;
      for (;;) {
        const next = [...timers].filter(([, timer]) => timer.at <= end).sort((a, b) => a[1].at - b[1].at)[0];
        if (!next) break;
        const [id, timer] = next;
        now = timer.at;
        timers.delete(id);
        timer.fn();
      }
      now = end;
    }
  };
}

test('startup failures retry with capped backoff, then join once on connection', () => {
  const f = setup();
  assert.equal(f.options.reconnection, false);
  assert.deepEqual(Array.from(f.options.transports), ['websocket']);
  assert.equal(f.options.timeout, 10000);
  for (const delay of [500, 1000, 2000, 2000]) {
    const before = f.connects;
    f.fire('connect_error');
    f.tick(delay - 1);
    assert.equal(f.connects, before);
    f.tick(1);
    assert.equal(f.connects, before + 1);
    assert.deepEqual(f.notifications, []);
    assert.deepEqual(f.sent, []);
  }
  f.fire('connect');
  f.fire('room_joined', { code: 'ABC234', role: 'PLAYER_1' });
  assert.equal(f.sent.length, 1);
  assert.equal(f.sent[0].event, 'join_room');
  assert.equal(f.sent[0].payload.hostToken, 'private-host-token');
  f.tick(60000);
  assert.equal(f.notifications.length, 1);
  assert.equal(f.notifications[0].type, 'room_joined');
  assert.equal(f.disconnects, 0);
  assert.equal(f.pendingTimers, 0);
});

test('unavailable or stalled signaling fails within the overall startup deadline', () => {
  for (const withErrors of [false, true]) {
    const f = setup();
    for (let i = 0; i < 30; i++) {
      if (withErrors) f.fire('connect_error');
      f.tick(1000);
    }
    assert.equal(f.notifications.length, 1);
    assert.equal(f.notifications[0].type, 'error');
    assert.match(f.notifications[0].data, /Signaling unavailable/);
    assert.equal(f.disconnects, 1);
    assert.equal(f.pendingTimers, 0);
    const before = f.connects;
    f.tick(60000);
    assert.equal(f.connects, before);
  }
});

test('leaving the page or closing the bridge cancels pending connection attempts', () => {
  for (const action of ['hide', 'close']) {
    const f = setup();
    f.fire('connect_error');
    if (action === 'hide') f.hide();
    else f.bridge.close();
    f.tick(60000);
    assert.equal(f.connects, 1);
    assert.equal(f.disconnects, 1);
    assert.equal(f.pendingTimers, 0);
    assert.deepEqual(f.notifications, []);
  }
});

test('an established signaling disconnect is terminal and never rejoins', () => {
  const f = setup();
  f.fire('connect');
  f.fire('room_joined', { code: 'ABC234', role: 'PLAYER_1' });
  f.fire('disconnect', 'transport error');
  f.tick(60000);
  assert.equal(f.connects, 1);
  assert.equal(f.sent.length, 1);
  assert.equal(f.notifications.at(-1).type, 'error');
  assert.match(f.notifications.at(-1).data, /Signaling disconnected/);
  assert.equal(f.pendingTimers, 0);
  assert.equal(f.warnings.length, 1);
  assert.equal(f.warnings[0][1].reason, 'transport error');
  assert.equal(f.warnings[0][1].transport, 'websocket');
  assert.doesNotMatch(JSON.stringify(f.warnings), /private-host-token|ABC234/);
});

test('a connected socket must confirm room admission before the startup deadline', () => {
  const f = setup();
  assert.deepEqual(f.statuses, ['Connecting to the game server…']);
  f.fire('connect');
  assert.equal(f.statuses.at(-1), 'Joining the room…');
  // A join acknowledgement alone must not leave the board waiting forever.
  f.tick(30000);
  assert.equal(f.notifications.length, 1);
  assert.equal(f.notifications[0].type, 'error');
  assert.match(f.notifications[0].data, /Room admission timed out/);
  assert.equal(f.disconnects, 1);
  assert.equal(f.pendingTimers, 0);
});

test('peer setup reports its stage and stops a stalled data channel', async () => {
  const f = setup();
  f.fire('connect');
  f.fire('room_joined', { code: 'ABC234', role: 'PLAYER_1' });
  f.fire('peer_joined', { code: 'ABC234' });
  await new Promise(resolve => setImmediate(resolve));
  assert.equal(f.statuses.at(-1), 'Connecting to your opponent…');
  assert.equal(f.sent.at(-1).event, 'signal_offer');
  assert.equal(f.channels.length, 1);
  f.tick(30000);
  assert.equal(f.notifications.at(-1).type, 'error');
  assert.match(f.notifications.at(-1).data, /TURN relay/);
  assert.equal(f.disconnects, 1);
  assert.equal(f.pendingTimers, 0);
});

test('an open data channel clears the peer setup deadline', async () => {
  const f = setup();
  f.fire('connect');
  f.fire('room_joined', { code: 'ABC234', role: 'PLAYER_1' });
  f.fire('peer_joined', { code: 'ABC234' });
  await new Promise(resolve => setImmediate(resolve));
  f.channels[0].onopen();
  f.tick(60000);
  assert.equal(f.notifications.at(-1).type, 'open');
  assert.equal(f.disconnects, 0);
  assert.equal(f.pendingTimers, 0);
});
