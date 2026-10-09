package shorten_url_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"

	"htqrcode"
	"htqrcode/shorten_url/adapters/db"
	"htqrcode/shorten_url/app"
	"htqrcode/shorten_url/domain"
)

func openDatabase(t *testing.T, path string) *sql.DB {
	t.Helper()
	database, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })
	return database
}

func newService(t *testing.T, database *sql.DB) htqrcode.Service {
	t.Helper()
	service, err := htqrcode.New(context.Background(), htqrcode.ExternalServices{Database: database})
	require.NoError(t, err)
	t.Cleanup(func() { service.GRPCServer().Stop() })
	return service
}

func create(handler http.Handler, key, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/shorten-url", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func TestAnonymousHTTPAndPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shorten.db")
	database := openDatabase(t, path)
	service := newService(t, database)
	response := create(service.HTTPHandler(), "request-1", `{"long_url":"https://example.com/a?b=1"}`)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	var result struct {
		ShortURL  string    `json:"short_url"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	require.Regexp(t, `^/r/[0-9A-Za-z]{8}$`, result.ShortURL)
	require.WithinDuration(t, time.Now().Add(30*24*time.Hour), result.ExpiresAt, 5*time.Second)
	replay := create(service.HTTPHandler(), "request-1", `{"long_url":"https://example.com/a?b=1"}`)
	require.Equal(t, response.Body.String(), replay.Body.String())
	// Anonymous callers do not share the free user's five-link quota.
	for i := 0; i < 6; i++ {
		created := create(service.HTTPHandler(), fmt.Sprintf("anonymous-%d", i), `{"long_url":"https://example.com"}`)
		require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	}
	require.NoError(t, database.Close())
	restartedDB := openDatabase(t, path)
	restarted := newService(t, restartedDB)
	redirect := httptest.NewRecorder()
	restarted.HTTPHandler().ServeHTTP(redirect, httptest.NewRequest(http.MethodGet, result.ShortURL, nil))
	require.Equal(t, http.StatusTemporaryRedirect, redirect.Code)
	require.Equal(t, "https://example.com/a?b=1", redirect.Header().Get("Location"))

	_, err := restartedDB.Exec("UPDATE shorten_url_links SET expires_at = ? WHERE short_code = ?", time.Now().Add(-time.Second).UnixNano(), strings.TrimPrefix(result.ShortURL, "/r/"))
	require.NoError(t, err)
	expired := httptest.NewRecorder()
	restarted.HTTPHandler().ServeHTTP(expired, httptest.NewRequest(http.MethodGet, result.ShortURL, nil))
	require.Equal(t, http.StatusNotFound, expired.Code)
	missing := httptest.NewRecorder()
	restarted.HTTPHandler().ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/r/00000000", nil))
	require.Equal(t, http.StatusNotFound, missing.Code)
}

func TestHTTPValidationAndSafeErrors(t *testing.T) {
	database := openDatabase(t, filepath.Join(t.TempDir(), "shorten.db"))
	service := newService(t, database)
	for _, test := range []struct{ key, body string }{
		{"", `{"long_url":"https://example.com"}`},
		{"   ", `{"long_url":"https://example.com"}`},
		{"invalid", `{"long_url":"ftp://example.com"}`},
		{"missing-url", `{}`},
		{"malformed", `{`},
	} {
		response := create(service.HTTPHandler(), test.key, test.body)
		require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
		var result map[string]any
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
		require.NotEmpty(t, result["message"])
		require.NotEmpty(t, result["slug"])
		require.Equal(t, []any{}, result["details"])
	}
	require.NoError(t, database.Close())
	response := create(service.HTTPHandler(), "database-failure", `{"long_url":"https://example.com"}`)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.NotContains(t, response.Body.String(), "database is closed")
	require.JSONEq(t, `{"message":"could not create short URL","slug":"shorten_url_creation_failed","details":[]}`, response.Body.String())
}

func TestConcurrentCreationAcrossSQLitePools(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shorten.db")
	first := openDatabase(t, path)
	_ = newService(t, first)
	second := openDatabase(t, path)
	services := []*app.Service{
		app.NewService(db.NewRepository(first), domain.RandomCodeGenerator{}),
		app.NewService(db.NewRepository(second), domain.RandomCodeGenerator{}),
	}
	userID := uuid.New()
	for _, sameKey := range []bool{true, false} {
		var wg sync.WaitGroup
		outcomes := make(chan error, 20)
		ids := make(chan uuid.UUID, 20)
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				input := app.CreateInput{LongURL: "https://example.com", IdempotencyKey: "shared"}
				if !sameKey {
					input.UserID, input.Plan, input.IdempotencyKey = userID, domain.Free, fmt.Sprintf("quota-%d", i)
				}
				output, err := services[i%2].Create(context.Background(), input)
				outcomes <- err
				if err == nil {
					ids <- output.URL.ID()
				}
			}(i)
		}
		wg.Wait()
		close(outcomes)
		close(ids)
		successes := 0
		for err := range outcomes {
			if err == nil {
				successes++
			} else {
				require.False(t, sameKey, "idempotent creation failed: %v", err)
				require.ErrorIs(t, err, app.ErrLiveURLLimit)
			}
		}
		if sameKey {
			require.Equal(t, 20, successes)
			var firstID uuid.UUID
			for id := range ids {
				if firstID == uuid.Nil {
					firstID = id
				}
				require.Equal(t, firstID, id)
			}
		} else {
			require.Equal(t, 5, successes)
		}
	}
}

func TestInactiveLinksAndCleanup(t *testing.T) {
	database := openDatabase(t, filepath.Join(t.TempDir(), "shorten.db"))
	_ = newService(t, database)
	repo := db.NewRepository(database)
	now := time.Now().UTC()
	for i, age := range []time.Duration{366 * 24 * time.Hour, 364 * 24 * time.Hour} {
		link, err := domain.NewShortURL(uuid.New(), uuid.Nil, "https://example.com", fmt.Sprintf("Ab%06d", i), fmt.Sprintf("cleanup-%d", i), now, 30*24*time.Hour)
		require.NoError(t, err)
		_, err = repo.Create(context.Background(), link, 0)
		require.NoError(t, err)
		_, err = database.Exec("UPDATE shorten_url_links SET deactivated_at = ? WHERE id = ?", now.Add(-age).UnixNano(), link.ID().String())
		require.NoError(t, err)
		_, err = repo.FindActiveByCode(context.Background(), link.Code(), now)
		require.ErrorIs(t, err, app.ErrNotFound)
	}
	require.NoError(t, app.NewService(repo, domain.RandomCodeGenerator{}).Cleanup(context.Background(), now))
	var count int
	require.NoError(t, database.QueryRow("SELECT count(*) FROM shorten_url_links").Scan(&count))
	require.Equal(t, 1, count)
}
