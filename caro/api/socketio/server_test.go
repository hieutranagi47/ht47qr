package socketio_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	"htqrcode/caro"
	"htqrcode/caro/adapters/db"
	caroHTTP "htqrcode/caro/api/http"
	"htqrcode/caro/api/socketio"
	"htqrcode/caro/app"
	"htqrcode/caro/domain"
	commonHTTP "htqrcode/common/http"
)

type fixture struct {
	server  *httptest.Server
	service *app.Service
	signals *socketio.Server
	now     *atomic.Int64
	cancel  context.CancelFunc
	done    chan error
	repo    app.Repository
}

func TestStartupAdmissionIsNotCacheable(t *testing.T) {
	e := commonHTTP.NewEcho()
	// Recovery has not run, so no service operation may be admitted yet.
	socketio.NewServer(nil).Register(e)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest("GET", socketio.Path+"?EIO=4&transport=polling", nil))
	require.Equal(t, 503, w.Code)
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
}

func setup(t *testing.T) fixture {
	t.Helper()
	pool, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "rooms.db")+"?_busy_timeout=5000")
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })
	require.NoError(t, caro.NewModule(pool).Init(context.Background()))
	clock := &atomic.Int64{}
	clock.Store(time.Now().UnixNano())
	repo := db.NewRepository(pool)
	service := app.NewService(repo, func() time.Time { return time.Unix(0, clock.Load()) }, domain.RandomCode)
	signals := socketio.NewServer(service)
	e := commonHTTP.NewEcho()
	signals.Register(e)
	require.NoError(t, caroHTTP.Register(context.Background(), e, caroHTTP.NewHandler(service)))
	server := httptest.NewServer(e)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- signals.Run(ctx) }()
	require.Eventually(t, func() bool {
		r, err := server.Client().Get(server.URL + socketio.Path + "?EIO=4&transport=polling")
		if err != nil {
			return false
		}
		defer r.Body.Close()
		_, _ = io.Copy(io.Discard, r.Body)
		return r.StatusCode == 200
	}, time.Second, 10*time.Millisecond)
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Error("signaling shutdown timed out")
		}
		server.Close()
	})
	return fixture{server, service, signals, clock, cancel, done, repo}
}

type wireClient struct {
	conn   *websocket.Conn
	events map[string][]json.RawMessage
	seq    int
}

func connect(t *testing.T, f fixture) *wireClient {
	t.Helper()
	url := "ws" + strings.TrimPrefix(f.server.URL, "http") + socketio.Path + "?EIO=4&transport=websocket"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	_, data, err := conn.ReadMessage()
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(data), "0"))
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte("40")))
	_, data, err = conn.ReadMessage()
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(data), "40"))
	return &wireClient{conn: conn, events: map[string][]json.RawMessage{}}
}

func (c *wireClient) emit(t *testing.T, event string, payload any) bool {
	t.Helper()
	c.seq++
	id := c.seq
	b, err := json.Marshal([]any{event, payload})
	require.NoError(t, err)
	require.NoError(t, c.conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("42%d%s", id, b))))
	for {
		require.NoError(t, c.conn.SetReadDeadline(time.Now().Add(5*time.Second)))
		_, b, err := c.conn.ReadMessage()
		require.NoError(t, err)
		packet := string(b)
		if packet == "2" {
			require.NoError(t, c.conn.WriteMessage(websocket.TextMessage, []byte("3")))
			continue
		}
		if strings.HasPrefix(packet, "42") {
			c.record(t, b[2:])
			continue
		}
		prefix := fmt.Sprintf("43%d", id)
		require.True(t, strings.HasPrefix(packet, prefix), packet)
		var values []struct {
			OK bool `json:"ok"`
		}
		require.NoError(t, json.Unmarshal(b[len(prefix):], &values))
		require.Len(t, values, 1)
		return values[0].OK
	}
}

func (c *wireClient) record(t *testing.T, b []byte) {
	t.Helper()
	var event []json.RawMessage
	require.NoError(t, json.Unmarshal(b, &event))
	require.Len(t, event, 2)
	var name string
	require.NoError(t, json.Unmarshal(event[0], &name))
	c.events[name] = append(c.events[name], event[1])
}

func (c *wireClient) event(t *testing.T, name string) json.RawMessage {
	t.Helper()
	for len(c.events[name]) == 0 {
		require.NoError(t, c.conn.SetReadDeadline(time.Now().Add(5*time.Second)))
		_, b, err := c.conn.ReadMessage()
		require.NoError(t, err)
		if string(b) == "2" {
			require.NoError(t, c.conn.WriteMessage(websocket.TextMessage, []byte("3")))
			continue
		}
		require.True(t, strings.HasPrefix(string(b), "42"), string(b))
		c.record(t, b[2:])
	}
	result := c.events[name][0]
	c.events[name] = c.events[name][1:]
	return result
}

func TestSocketIOMembershipRelayAndExpiry(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	created, err := f.service.Create(ctx, app.CreateInput{IdempotencyKey: uuid.NewString(), DisplayName: "Alice", RoomType: "PUBLIC", MaxGames: 3, OwnerIP: "192.0.2.1"})
	require.NoError(t, err)
	code := created.Room.Code()
	host := connect(t, f)
	guest := connect(t, f)
	outsider := connect(t, f)
	require.False(t, outsider.emit(t, "join_room", map[string]any{"code": code, "hostToken": strings.Repeat("x", 64)}))
	require.True(t, guest.emit(t, "join_room", map[string]any{"code": strings.ToLower(code), "displayName": "Bob"}))
	require.NotEmpty(t, guest.event(t, "room_joined"))
	room, err := f.service.Get(ctx, code)
	require.NoError(t, err)
	require.Equal(t, domain.Waiting, room.Status())
	require.True(t, host.emit(t, "join_room", map[string]any{"code": code, "hostToken": created.HostToken}))
	require.NotEmpty(t, host.event(t, "room_joined"))
	require.NotEmpty(t, host.event(t, "peer_joined"))
	require.True(t, host.emit(t, "join_room", map[string]any{"code": code, "hostToken": created.HostToken}))
	require.Empty(t, host.events["room_joined"])
	require.Empty(t, host.events["peer_joined"])
	require.False(t, outsider.emit(t, "join_room", map[string]any{"code": code, "displayName": "Eve"}))
	offer := map[string]any{"code": code, "data": map[string]any{"type": "offer", "sdp": "test-sdp"}}
	require.False(t, outsider.emit(t, "signal_offer", offer))
	require.True(t, host.emit(t, "signal_offer", offer))
	require.JSONEq(t, `{"code":"`+code+`","data":{"type":"offer","sdp":"test-sdp"}}`, string(guest.event(t, "signal_offer")))
	require.False(t, host.emit(t, "signal_offer", map[string]any{"code": code, "data": map[string]any{"type": "answer", "sdp": "bad"}}))
	require.False(t, host.emit(t, "signal_ice", map[string]any{"code": "ZZZZZZ", "data": map[string]any{"candidate": "ice"}}))
	require.True(t, guest.emit(t, "signal_answer", map[string]any{"code": code, "data": map[string]any{"type": "answer", "sdp": "answer"}}))
	require.NotEmpty(t, host.event(t, "signal_answer"))
	f.now.Add(int64(domain.Lifetime))
	require.False(t, host.emit(t, "signal_offer", offer))
	cleaner := app.NewCleaner(f.repo, func() time.Time { return time.Unix(0, f.now.Load()) }, f.signals)
	count, err := cleaner.Cleanup(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 1, count)
	count, err = cleaner.Cleanup(ctx)
	require.NoError(t, err)
	require.Zero(t, count)
	// The Socket.IO disconnect packet may arrive before the WebSocket closes.
	for range 10 {
		_, _, err = host.conn.ReadMessage()
		if err != nil {
			break
		}
	}
	require.Error(t, err)
}

func TestSocketIOLeaveReconnectAndOrigin(t *testing.T) {
	f := setup(t)
	created, err := f.service.Create(context.Background(), app.CreateInput{IdempotencyKey: uuid.NewString(), DisplayName: "Alice", RoomType: "PRIVATE", MaxGames: 1, OwnerIP: "192.0.2.1"})
	require.NoError(t, err)
	code := created.Room.Code()
	host := connect(t, f)
	require.True(t, host.emit(t, "join_room", map[string]any{"code": code, "hostToken": created.HostToken}))
	require.NotEmpty(t, host.event(t, "room_joined"))
	require.True(t, host.emit(t, "leave_room", map[string]any{"code": code}))
	require.True(t, host.emit(t, "leave_room", map[string]any{"code": code}))
	replacement := connect(t, f)
	require.True(t, replacement.emit(t, "join_room", map[string]any{"code": code, "hostToken": created.HostToken}))
	host.conn.Close()
	guest := connect(t, f)
	require.True(t, guest.emit(t, "join_room", map[string]any{"code": code, "displayName": "Bob"}))
	require.NotEmpty(t, replacement.event(t, "peer_joined"))
	require.True(t, guest.emit(t, "leave_room", map[string]any{"code": code}))
	require.NotEmpty(t, replacement.event(t, "peer_disconnected"))
	require.True(t, guest.emit(t, "leave_room", map[string]any{"code": code}))
	require.True(t, guest.emit(t, "join_room", map[string]any{"code": code, "displayName": "Bob"}))
	require.NotEmpty(t, replacement.event(t, "peer_joined"))
	url := "ws" + strings.TrimPrefix(f.server.URL, "http") + socketio.Path + "?EIO=4&transport=websocket"
	_, response, err := websocket.DefaultDialer.Dial(url, map[string][]string{"Origin": {"https://evil.example"}})
	require.Error(t, err)
	require.Equal(t, 403, response.StatusCode)
	response.Body.Close()
}

func TestSocketIOMultipleActiveRoomsPerIP(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	input := app.CreateInput{IdempotencyKey: uuid.NewString(), DisplayName: "Host", RoomType: "PUBLIC", MaxGames: 1, OwnerIP: "192.0.2.1"}
	active, err := f.service.Create(ctx, input)
	require.NoError(t, err)
	host := connect(t, f)
	guest := connect(t, f)
	require.True(t, host.emit(t, "join_room", map[string]any{"code": active.Room.Code(), "hostToken": active.HostToken}))
	require.True(t, guest.emit(t, "join_room", map[string]any{"code": active.Room.Code(), "displayName": "Guest"}))
	input.IdempotencyKey = uuid.NewString()
	waiting, err := f.service.Create(ctx, input)
	require.NoError(t, err)
	waitingHost := connect(t, f)
	waitingGuest := connect(t, f)
	require.True(t, waitingHost.emit(t, "join_room", map[string]any{"code": waiting.Room.Code(), "hostToken": waiting.HostToken}))
	require.True(t, waitingGuest.emit(t, "join_room", map[string]any{"code": waiting.Room.Code(), "displayName": "Next guest"}))
	stored, err := f.service.Get(ctx, waiting.Room.Code())
	require.NoError(t, err)
	require.Equal(t, domain.Playing, stored.Status())
	require.Len(t, stored.Players(), 2)
	require.NotEmpty(t, waitingHost.event(t, "peer_joined"))
}

func TestSocketIOGuestIPRoomLimit(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	create := func(ip string) app.Created {
		r, err := f.service.Create(ctx, app.CreateInput{DisplayName: "Host", RoomType: "PUBLIC", MaxGames: 1, OwnerIP: ip, IdempotencyKey: uuid.NewString()})
		require.NoError(t, err)
		return r
	}
	for range domain.MaxRoomsPerIP - 1 {
		create("127.0.0.1")
	}
	first := create("192.0.2.1")
	second := create("192.0.2.2")
	a := connect(t, f)
	b := connect(t, f)
	require.True(t, a.emit(t, "join_room", map[string]any{"code": first.Room.Code(), "displayName": "Guest"}))
	// Both WebSocket handshakes resolve to the same loopback IP.
	require.False(t, b.emit(t, "join_room", map[string]any{"code": second.Room.Code(), "displayName": "Guest"}))
	require.Contains(t, string(b.event(t, "room_error")), "waiting_room_limit_reached")
	require.True(t, a.emit(t, "leave_room", map[string]any{"code": first.Room.Code()}))
	require.True(t, b.emit(t, "join_room", map[string]any{"code": second.Room.Code(), "displayName": "Guest"}))
}

// Refresh replaces both transport sockets without sending a leave_room event.
func TestSocketIOBothPlayersRefreshAndRejoin(t *testing.T) {
	for _, hostFirst := range []bool{true, false} {
		t.Run(fmt.Sprintf("host_first_%t", hostFirst), func(t *testing.T) {
			f := setup(t)
			ctx := context.Background()
			created, err := f.service.Create(ctx, app.CreateInput{IdempotencyKey: uuid.NewString(), DisplayName: "Alice", RoomType: "PUBLIC", MaxGames: 3, OwnerIP: "192.0.2.1"})
			require.NoError(t, err)
			code := created.Room.Code()
			host, guest := connect(t, f), connect(t, f)
			require.True(t, host.emit(t, "join_room", map[string]any{"code": code, "hostToken": created.HostToken}))
			require.True(t, guest.emit(t, "join_room", map[string]any{"code": code, "displayName": "Bob"}))
			host.conn.Close()
			guest.conn.Close()
			require.Eventually(t, func() bool {
				r, err := f.service.Get(ctx, code)
				return err == nil && r.Status() == domain.Waiting && len(r.Players()) == 1 && r.Players()[0].Socket() == ""
			}, time.Second, 10*time.Millisecond)
			newHost, newGuest := connect(t, f), connect(t, f)
			joinHost := func() {
				require.True(t, newHost.emit(t, "join_room", map[string]any{"code": code, "hostToken": created.HostToken}))
			}
			joinGuest := func() {
				require.True(t, newGuest.emit(t, "join_room", map[string]any{"code": code, "displayName": "Bob"}))
			}
			if hostFirst {
				joinHost()
				joinGuest()
			} else {
				joinGuest()
				joinHost()
			}
			require.NotEmpty(t, newHost.event(t, "peer_joined"))
			r, err := f.service.Get(ctx, code)
			require.NoError(t, err)
			require.Equal(t, domain.Playing, r.Status())
			require.Equal(t, created.Room.ExpiresAt(), r.ExpiresAt())
			require.True(t, newHost.emit(t, "signal_offer", map[string]any{"code": code, "data": map[string]any{"type": "offer", "sdp": "fresh-offer"}}))
			require.NotEmpty(t, newGuest.event(t, "signal_offer"))
		})
	}
}
