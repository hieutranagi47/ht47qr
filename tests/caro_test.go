package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"htqrcode"
)

func TestCaroRESTComponent(t *testing.T) {
	for _, backend := range []string{"sqlite", "postgres"} {
		t.Run(backend, func(t *testing.T) {
			ctx := context.Background()
			dependencies := htqrcode.ExternalServices{}
			if backend == "sqlite" {
				db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "service.db")+"?_busy_timeout=5000")
				require.NoError(t, err)
				t.Cleanup(func() { db.Close() })
				dependencies.Database = db
			} else {
				url := os.Getenv("CARO_TEST_POSTGRES_URL")
				if url == "" {
					t.Skip("CARO_TEST_POSTGRES_URL not configured")
				}
				pool, err := pgxpool.New(ctx, url)
				require.NoError(t, err)
				t.Cleanup(pool.Close)
				lock, err := pool.Acquire(ctx)
				require.NoError(t, err)
				_, err = lock.Exec(ctx, "SELECT pg_advisory_lock(47470001)")
				require.NoError(t, err)
				t.Cleanup(func() { _, _ = lock.Exec(context.Background(), "SELECT pg_advisory_unlock(47470001)"); lock.Release() })
				dependencies.Postgres = pool
			}
			service, err := htqrcode.New(ctx, dependencies)
			require.NoError(t, err)
			if dependencies.Postgres != nil {
				_, err = dependencies.Postgres.Exec(ctx, "TRUNCATE caro.rooms CASCADE")
				require.NoError(t, err)
				t.Cleanup(func() { _, _ = dependencies.Postgres.Exec(context.Background(), "TRUNCATE caro.rooms CASCADE") })
			}
			handler := service.HTTPHandler()
			requestIndex := 0
			call := func(method, path, body string) *httptest.ResponseRecorder {
				req := httptest.NewRequest(method, path, strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Idempotency-Key", uuid.NewString())
				requestIndex++
				req.RemoteAddr = fmt.Sprintf("192.0.2.%d:1234", requestIndex)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				return w
			}
			for _, games := range []int{1, 3, 5, 7} {
				w := call("POST", "/api/v1/rooms", fmt.Sprintf(`{"displayName":"Alice","roomType":"PRIVATE","maxGames":%d}`, games))
				require.Equal(t, 201, w.Code, w.Body.String())
				var created struct {
					Room struct {
						Code      string `json:"code"`
						Occupancy int    `json:"occupancy"`
						Joinable  bool   `json:"joinable"`
					}
					HostReservation struct {
						HostToken string `json:"hostToken"`
					}
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
				require.Len(t, created.Room.Code, 6)
				require.Len(t, created.HostReservation.HostToken, 64)
				require.Equal(t, 1, created.Room.Occupancy)
				require.True(t, created.Room.Joinable)
				w = call("GET", "/api/v1/rooms/"+strings.ToLower(created.Room.Code), "")
				require.Equal(t, 200, w.Code)
				require.NotContains(t, w.Body.String(), created.HostReservation.HostToken)
				require.NotContains(t, w.Body.String(), "socket")
				require.NotContains(t, w.Body.String(), "credential")
			}
			w := call("GET", "/api/v1/rooms", "")
			require.Equal(t, 200, w.Code)
			require.JSONEq(t, `{"rooms":[]}`, w.Body.String())
			for _, body := range []string{"", `{`, `null`, `{}`, `{"displayName":"","roomType":"PUBLIC","maxGames":1}`, `{"displayName":"Alice","roomType":"PUBLIC","maxGames":2}`, `{"displayName":"Alice","roomType":"OTHER","maxGames":1}`, `{"displayName":42,"roomType":"PUBLIC","maxGames":1}`} {
				w = call("POST", "/api/v1/rooms", body)
				require.Equal(t, 400, w.Code, w.Body.String())
			}
			w = call("GET", "/api/v1/rooms/BAD", "")
			require.Equal(t, 400, w.Code)
			w = call("GET", "/api/v1/rooms/ZZZZZZ", "")
			require.Equal(t, 404, w.Code)
			w = call("POST", "/api/v1/rooms", `{"displayName":"Bob","roomType":"PUBLIC","maxGames":3}`)
			require.Equal(t, 201, w.Code)
			w = call("GET", "/api/v1/rooms", "")
			require.Equal(t, 200, w.Code)
			require.NotContains(t, w.Body.String(), "PRIVATE")
			require.Contains(t, w.Body.String(), "PUBLIC")
			for path, status := range map[string]int{"/health": 200, "/api/qrcode/v1/very-first-api": 200, "/": 200, "/r/missing": 404} {
				w = call("GET", path, "")
				require.Equal(t, status, w.Code, path)
			}
			// Same-origin preflight uses the existing shared CORS policy.
			req := httptest.NewRequest("OPTIONS", "/api/v1/rooms", nil)
			req.Header.Set("Origin", "http://example.com")
			req.Header.Set("Access-Control-Request-Method", "POST")
			w = httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			require.Equal(t, 204, w.Code)
			require.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
			req = httptest.NewRequest("OPTIONS", "/api/v1/rooms", nil)
			req.Header.Set("Origin", "https://evil.example")
			req.Header.Set("Access-Control-Request-Method", "POST")
			w = httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			require.Equal(t, 403, w.Code)
			bg, cancel := context.WithCancel(ctx)
			done := make(chan error, 1)
			go func() { done <- service.RunBackground(bg) }()
			// Wait for signaling admission, then verify service cancellation owns shutdown.
			server := httptest.NewServer(handler)
			defer server.Close()
			require.Eventually(t, func() bool {
				r, err := server.Client().Get(server.URL + "/api/v1/caro/socket.io/?EIO=4&transport=polling")
				if err != nil {
					return false
				}
				defer r.Body.Close()
				_, _ = io.Copy(io.Discard, r.Body)
				return r.StatusCode == 200
			}, 5*time.Second, 50*time.Millisecond)
			cancel()
			select {
			case err := <-done:
				require.NoError(t, err)
			case <-time.After(5 * time.Second):
				t.Fatal("background shutdown timed out")
			}
		})
	}
}

func TestCaroOmittedWithoutDatabase(t *testing.T) {
	service, err := htqrcode.New(context.Background(), htqrcode.ExternalServices{})
	require.NoError(t, err)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", uuid.NewString())
	service.HTTPHandler().ServeHTTP(w, req)
	require.Equal(t, 405, w.Code)
}

func TestCaroBrowserSocketIOClient(t *testing.T) {
	if os.Getenv("CARO_SOCKETIO_CLIENT_DIR") == "" {
		t.Skip("set CARO_SOCKETIO_CLIENT_DIR to a temporary npm project with socket.io-client@4.8.1")
	}
	pool, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "client.db")+"?_busy_timeout=5000")
	require.NoError(t, err)
	defer pool.Close()
	service, err := htqrcode.New(context.Background(), htqrcode.ExternalServices{Database: pool})
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- service.RunBackground(ctx) }()
	server := httptest.NewServer(service.HTTPHandler())
	defer server.Close()
	defer func() {
		cancel()
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Error("shutdown timed out")
		}
	}()
	commandCtx, stop := context.WithTimeout(context.Background(), 20*time.Second)
	defer stop()
	command := exec.CommandContext(commandCtx, "node", "testdata/caro_socketio.cjs", server.URL)
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
	t.Log(string(output))
}

func TestCaroCleanupOwnedByBackground(t *testing.T) {
	pool, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "cleanup.db")+"?_busy_timeout=5000")
	require.NoError(t, err)
	defer pool.Close()
	service, err := htqrcode.New(context.Background(), htqrcode.ExternalServices{Database: pool})
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "/api/v1/rooms", strings.NewReader(`{"displayName":"Alice","roomType":"PUBLIC","maxGames":1}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", uuid.NewString())
	w := httptest.NewRecorder()
	service.HTTPHandler().ServeHTTP(w, req)
	require.Equal(t, 201, w.Code)
	past := time.Now().UTC().Add(-48 * time.Hour)
	_, err = pool.Exec("UPDATE caro_rooms SET created_at=?,expires_at=?", past.UnixNano(), past.Add(24*time.Hour).UnixNano())
	require.NoError(t, err)
	var count int
	require.NoError(t, pool.QueryRow("SELECT count(*) FROM caro_rooms").Scan(&count))
	require.Equal(t, 1, count)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- service.RunBackground(ctx) }()
	require.Eventually(t, func() bool {
		err := pool.QueryRow("SELECT count(*) FROM caro_rooms").Scan(&count)
		return err == nil && count == 0
	}, time.Second, 10*time.Millisecond)
	cancel()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
	require.NoError(t, pool.QueryRow("SELECT count(*) FROM caro_room_players").Scan(&count))
	require.Zero(t, count)
	require.NoError(t, pool.Ping())
}

func TestCaroOwnerIPQuotaHTTP(t *testing.T) {
	pool, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "quota.db")+"?_busy_timeout=5000")
	require.NoError(t, err)
	defer pool.Close()
	service, err := htqrcode.New(context.Background(), htqrcode.ExternalServices{Database: pool})
	require.NoError(t, err)
	create := func(address, forwarded, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/api/v1/rooms", strings.NewReader(body))
		req.RemoteAddr = address
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", uuid.NewString())
		req.Header.Set("X-Forwarded-For", forwarded)
		req.Header.Set("X-Real-IP", forwarded)
		w := httptest.NewRecorder()
		service.HTTPHandler().ServeHTTP(w, req)
		return w
	}
	body := `{"displayName":"Alice","roomType":"PUBLIC","maxGames":1}`
	first := create("192.0.2.1:1234", "198.51.100.1", body)
	require.Equal(t, 201, first.Code, first.Body.String())
	duplicate := create("192.0.2.1:5678", "198.51.100.2", body)
	require.Equal(t, 409, duplicate.Code)
	require.Contains(t, duplicate.Body.String(), "waiting_room_limit_reached")
	mapped := create("[::ffff:192.0.2.1]:5678", "", body)
	require.Equal(t, 409, mapped.Code)
	other := create("192.0.2.2:1234", "198.51.100.1", body)
	require.Equal(t, 201, other.Code, other.Body.String())
	spoofed := create("192.0.2.1:1234", "", `{"displayName":"Eve","roomType":"PUBLIC","maxGames":1,"ownerIP":"192.0.2.3"}`)
	require.Equal(t, 400, spoofed.Code)
	require.NotContains(t, first.Body.String(), "192.0.2.1")
	require.NotContains(t, first.Body.String(), "ownerIP")
}

// The first response is lost after the database transaction commits.
type interruptedRoomResponse struct{ header http.Header }

func (w interruptedRoomResponse) Header() http.Header     { return w.header }
func (interruptedRoomResponse) WriteHeader(int)           {}
func (interruptedRoomResponse) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestCaroCreationRetryAfterInterruptedResponse(t *testing.T) {
	pool, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "retry.db")+"?_busy_timeout=5000")
	require.NoError(t, err)
	defer pool.Close()
	service, err := htqrcode.New(context.Background(), htqrcode.ExternalServices{Database: pool})
	require.NoError(t, err)
	key := uuid.NewString()
	body := `{"displayName":"Alice","roomType":"PRIVATE","maxGames":3}`
	request := func(key, body string) *http.Request {
		req := httptest.NewRequest("POST", "/api/v1/rooms", strings.NewReader(body))
		req.RemoteAddr = "192.0.2.1:1234"
		req.Header.Set("Content-Type", "application/json")
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		return req
	}
	service.HTTPHandler().ServeHTTP(interruptedRoomResponse{header: http.Header{}}, request(key, body))
	var count int
	require.NoError(t, pool.QueryRow("SELECT count(*) FROM caro_rooms").Scan(&count))
	require.Equal(t, 1, count)
	// A fresh service has no process-local response cache or secret from the first call.
	restarted, err := htqrcode.New(context.Background(), htqrcode.ExternalServices{Database: pool})
	require.NoError(t, err)
	retry := httptest.NewRecorder()
	restarted.HTTPHandler().ServeHTTP(retry, request(key, body))
	require.Equal(t, 201, retry.Code, retry.Body.String())
	require.Equal(t, "no-store", retry.Header().Get("Cache-Control"))
	again := httptest.NewRecorder()
	restarted.HTTPHandler().ServeHTTP(again, request(strings.ToUpper(key), body))
	require.Equal(t, 201, again.Code)
	require.Equal(t, retry.Body.String(), again.Body.String())
	var response struct {
		Room            struct{ Code string }
		HostReservation struct{ HostToken string }
	}
	require.NoError(t, json.Unmarshal(retry.Body.Bytes(), &response))
	require.Len(t, response.HostReservation.HostToken, 64)
	var code, keyHash, credentialHash string
	require.NoError(t, pool.QueryRow("SELECT code,idempotency_key_hash FROM caro_rooms").Scan(&code, &keyHash))
	require.Equal(t, code, response.Room.Code)
	require.NotEqual(t, key, keyHash)
	require.NoError(t, pool.QueryRow("SELECT credential_hash FROM caro_room_players").Scan(&credentialHash))
	require.NotEqual(t, response.HostReservation.HostToken, credentialHash)
	conflict := httptest.NewRecorder()
	restarted.HTTPHandler().ServeHTTP(conflict, request(key, `{"displayName":"Alice","roomType":"PRIVATE","maxGames":7}`))
	require.Equal(t, 409, conflict.Code)
	require.Contains(t, conflict.Body.String(), "idempotency_conflict")
	for _, invalidKey := range []string{"", "invalid", uuid.NewSHA1(uuid.Nil, []byte("non-v4")).String()} {
		w := httptest.NewRecorder()
		restarted.HTTPHandler().ServeHTTP(w, request(invalidKey, body))
		require.Equal(t, 400, w.Code)
	}
	require.NoError(t, pool.QueryRow("SELECT count(*) FROM caro_rooms").Scan(&count))
	require.Equal(t, 1, count)
}
