const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const { join } = require('node:path');
const { test } = require('node:test');
const vm = require('node:vm');

const source = readFileSync(join(__dirname, '../../caro/api/wasm/app.js'), 'utf8');
const settle = () => new Promise(resolve => setImmediate(resolve));

function setup(status = 'PLAYING') {
  const elements = new Map(), timers = new Map(), events = new Map();
  let requests = 0, nextID = 0;
  const room = { code: 'ABC234', maxGames: 3, roomType: 'PUBLIC', occupancy: 2, status, joinable: false };
  const element = id => {
    if (!elements.has(id)) elements.set(id, { value: '', disabled: false, parentElement: {} });
    return elements.get(id);
  };
  vm.runInNewContext(source, {
    document: { getElementById: element },
    sessionStorage: { getItem() { return null; } },
    location: { pathname: '/caro/boards/ABC234' },
    window: { addEventListener(event, fn) { events.set(event, fn); } },
    async fetch() { requests++; return { ok: true, async json() { return { ...room }; } }; },
    setTimeout(fn) { const id = ++nextID; timers.set(id, fn); return id; },
    clearTimeout(id) { timers.delete(id); }
  });
  return {
    room, element, timers,
    get requests() { return requests; },
    retry() { const [id, fn] = timers.entries().next().value; timers.delete(id); fn(); },
    hide() { events.get('pagehide')(); }
  };
}

test('a guest refreshing before its old socket leaves can enter once the seat opens', async () => {
  const f = setup();
  await settle();
  assert.equal(f.element('enter-button').disabled, true);
  assert.equal(f.timers.size, 1);
  f.room.status = 'WAITING';
  f.room.joinable = true;
  f.room.occupancy = 1;
  f.retry();
  await settle();
  assert.equal(f.requests, 2);
  assert.equal(f.element('enter-button').disabled, false);
  assert.equal(f.element('room-status').textContent, '');
  assert.equal(f.timers.size, 0);
});

test('leaving a full room page cancels its admission retry', async () => {
  const f = setup();
  await settle();
  f.hide();
  assert.equal(f.timers.size, 0);
});

test('a finished room remains closed without retries', async () => {
  const f = setup('FINISHED');
  await settle();
  assert.equal(f.element('enter-button').disabled, true);
  assert.match(f.element('room-status').textContent, /finished/);
  assert.equal(f.timers.size, 0);
});
