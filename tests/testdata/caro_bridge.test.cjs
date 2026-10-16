const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const { join } = require('node:path');
const { test } = require('node:test');
const vm = require('node:vm');

const source = readFileSync(join(__dirname, '../../caro/api/wasm/bridge.js'), 'utf8');

function setup() {
  let now = 0, nextID = 0, options, connects = 0, disconnects = 0;
  const timers = new Map(), handlers = new Map(), pageEvents = new Map();
  const notifications = [], sent = [];
  const socket = {
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
    io(opts) { options = opts; return socket; },
    setTimeout(fn, delay) { const id = ++nextID; timers.set(id, { fn, at: now + delay }); return id; },
    clearTimeout(id) { timers.delete(id); }
  });
  const bridge = window.caroBridge.connect((type, data) => notifications.push({ type, data }));
  return {
    bridge, notifications, sent, options,
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
  f.fire('disconnect');
  f.tick(60000);
  assert.equal(f.connects, 1);
  assert.equal(f.sent.length, 1);
  assert.equal(f.notifications.at(-1).type, 'error');
  assert.match(f.notifications.at(-1).data, /Signaling disconnected/);
  assert.equal(f.pendingTimers, 0);
});
