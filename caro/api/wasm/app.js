'use strict';
(() => {
  const $ = id => document.getElementById(id);
  const storage = {
    get(key) { try { return JSON.parse(sessionStorage.getItem('caro:' + key)); } catch { return null; } },
    set(key, value) { sessionStorage.setItem('caro:' + key, JSON.stringify(value)); },
    remove(key) { sessionStorage.removeItem('caro:' + key); }
  };
  async function request(path, options) {
    const response = await fetch(path, { ...options, cache: 'no-store' });
    const body = await response.json().catch(() => ({}));
    if (!response.ok) {
      const error = new Error(body.message || body.error?.message || `Request failed (${response.status})`);
      error.status = response.status;
      throw error;
    }
    return body;
  }
  const name = storage.get('name') || '';
  $('display-name').value = name;
  $('guest-name').value = name;
  let refreshing = false;
  async function refresh() {
    if (refreshing) return;
    refreshing = true;
    $('refresh').disabled = true;
    try {
      const { rooms } = await request('/api/v1/rooms');
      $('room-count').textContent = rooms.length;
      const cards = rooms.map(room => {
        const card = document.createElement('article'); card.className = 'room-card';
        const info = document.createElement('div');
        const title = document.createElement('h3'); title.textContent = room.players[0]?.displayName || 'Open room';
        const detail = document.createElement('p'); detail.textContent = `${room.code} · Best of ${room.maxGames} · ${room.occupancy}/2 players`;
        const link = document.createElement('a'); link.href = '/caro/boards/' + encodeURIComponent(room.code); link.textContent = 'Take a seat ↗';
        info.append(title, detail); card.append(info, link); return card;
      });
      if (!cards.length) { const empty = document.createElement('p'); empty.className = 'empty'; empty.textContent = 'A fresh board is waiting. Create the first room.'; cards.push(empty); }
      $('rooms').replaceChildren(...cards);
    } catch (error) { $('status').textContent = error.message; }
    finally { refreshing = false; $('refresh').disabled = false; }
  }
  $('refresh').onclick = refresh;
  // Keep the private key + exact input until a creation receipt is safely stored.
  const pending = storage.get('creation');
  if (pending) {
    const form = $('create-form');
    form.elements.displayName.value = pending.body.displayName;
    form.elements.maxGames.value = pending.body.maxGames;
    form.elements.roomType.value = pending.body.roomType;
    $('status').textContent = 'A room request was interrupted. Create a room again to recover it.';
  }
  $('create-form').onsubmit = async event => {
    event.preventDefault();
    const button = $('create-button'); button.disabled = true; $('status').textContent = '';
    try {
      const form = event.currentTarget;
      let attempt = storage.get('creation');
      if (!attempt) {
        const displayName = form.elements.displayName.value.trim();
        if (!displayName) throw new Error('Enter your name.');
        attempt = { key: crypto.randomUUID(), body: { displayName, maxGames: Number(form.elements.maxGames.value), roomType: form.elements.roomType.value } };
        storage.set('creation', attempt);
      }
      const result = await request('/api/v1/rooms', { method: 'POST', headers: { 'Content-Type': 'application/json', 'Idempotency-Key': attempt.key }, body: JSON.stringify(attempt.body) });
      storage.set('host:' + result.room.code, result.hostReservation);
      storage.set('name', attempt.body.displayName);
      storage.remove('creation');
      location.assign('/caro/boards/' + result.room.code);
    } catch (error) {
      // A received 4xx is definitive; network/5xx failures must reuse the key.
      if (error.status >= 400 && error.status < 500) storage.remove('creation');
      $('status').textContent = error.message + (storage.get('creation') ? ' Submit again to retry the same request.' : '');
    } finally { button.disabled = false; }
  };
  $('join-form').onsubmit = event => {
    event.preventDefault();
    const value = $('display-name').value.trim();
    if (value) storage.set('name', value);
    const code = event.currentTarget.elements.code.value.trim().toUpperCase();
    if (/^[A-HJ-NP-Z2-9]{6}$/.test(code)) location.assign('/caro/boards/' + code);
  };
  const match = location.pathname.match(/^\/caro\/boards\/([^/]+)\/?$/);
  if (!match) {
    refresh();
    const interval = setInterval(() => { if (!document.hidden) refresh(); }, 15000);
    window.addEventListener('pagehide', () => clearInterval(interval), { once: true });
    return;
  }
  $('lobby').hidden = true; $('room').hidden = false;
  let code;
  try { code = decodeURIComponent(match[1]).trim().toUpperCase(); } catch { code = ''; }
  let room, starting = false;
  const reservation = storage.get('host:' + code);
  $('copy-link').onclick = async () => {
    try { await navigator.clipboard.writeText(location.origin + '/caro/boards/' + encodeURIComponent(code)); $('room-status').textContent = 'Invite link copied.'; }
    catch { $('room-status').textContent = 'Share the URL in your address bar.'; }
  };
  async function loadRoom() {
    try {
      if (!/^[A-HJ-NP-Z2-9]{6}$/.test(code)) throw new Error('Enter a valid six-character room code.');
      room = await request('/api/v1/rooms/' + code);
      $('room-title').textContent = 'Board ' + room.code;
      $('room-description').textContent = `Best of ${room.maxGames} · ${room.roomType.toLowerCase()} · ${room.occupancy}/2 seats reserved`;
      if (room.status === 'FINISHED' || (!reservation && !room.joinable)) throw new Error('This room is full or finished. Return to rooms to find another board.');
      if (reservation) {
        $('guest-name').required = false; $('guest-name').parentElement.hidden = true;
        $('enter-button').textContent = 'Open your board ↗';
      }
    } catch (error) { $('room-status').textContent = error.message; $('enter-button').disabled = true; }
  }
  $('enter-form').onsubmit = async event => {
    event.preventDefault(); if (starting || !room) return;
    const displayName = $('guest-name').value.trim();
    if (!reservation && !displayName) { $('room-status').textContent = 'Enter your name.'; return; }
    starting = true; $('enter-button').disabled = true; $('room-status').textContent = 'Loading game…';
    try {
      storage.set('name', displayName || name);
      window.caroConfig = { code, maxGames: room.maxGames, displayName, hostToken: reservation?.hostToken || '' };
      if (typeof Go !== 'function' || typeof io !== 'function' || !window.caroBridge) throw new Error('Game scripts could not load. Reload this page.');
      if (!window.RTCPeerConnection) throw new Error('WebRTC is unavailable in this browser.');
      const go = new Go();
      const response = await fetch('/caro/wasm/caro.wasm');
      if (!response.ok) throw new Error('Game build unavailable. Run the Caro build-wasm task.');
      // Buffer compilation also works behind proxies with incorrect WASM MIME types.
      const result = await WebAssembly.instantiate(await response.arrayBuffer(), go.importObject);
      document.body.classList.add('game-running');
      await go.run(result.instance);
    } catch (error) {
      document.body.classList.remove('game-running');
      $('room-status').textContent = error.message; starting = false; $('enter-button').disabled = false;
    }
  };
  loadRoom();
})();
