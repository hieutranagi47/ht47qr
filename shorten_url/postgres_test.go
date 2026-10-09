package shorten_url_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"htqrcode"
	"htqrcode/shorten_url/adapters/postgres"
	"htqrcode/shorten_url/app"
	"htqrcode/shorten_url/domain"
)

// Use a dedicated disposable database: this test resets the shorten_url tables.
func TestPostgresShortenURL(t *testing.T) {
	dsn := os.Getenv("SHORTEN_URL_TEST_POSTGRES_URL")
	if dsn == "" {
		t.Skip("set SHORTEN_URL_TEST_POSTGRES_URL to a disposable PostgreSQL database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	first, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer first.Close()
	second, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer second.Close()
	newService := func(pool *pgxpool.Pool) htqrcode.Service {
		service, err := htqrcode.New(ctx, htqrcode.ExternalServices{Postgres: pool, PostgresMigrations: second})
		require.NoError(t, err)
		t.Cleanup(func() { service.GRPCServer().Stop() })
		return service
	}
	service := newService(first)
	_, err = first.Exec(ctx, "TRUNCATE shorten_url.links CASCADE")
	require.NoError(t, err)
	// Repeated migration initialization must release all acquired connections.
	require.Zero(t, second.Stat().AcquiredConns())
	response := create(service.HTTPHandler(), "http", `{"long_url":"https://example.com/a?b=1"}`)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	replay := create(service.HTTPHandler(), "http", `{"long_url":"https://example.com/a?b=1"}`)
	require.Equal(t, response.Body.String(), replay.Body.String())
	var output struct {
		ShortURL  string    `json:"short_url"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &output))
	require.WithinDuration(t, time.Now().Add(7*24*time.Hour), output.ExpiresAt, 5*time.Second)
	restarted := newService(second)
	redirect := httptest.NewRecorder()
	restarted.HTTPHandler().ServeHTTP(redirect, httptest.NewRequest(http.MethodGet, output.ShortURL, nil))
	require.Equal(t, http.StatusTemporaryRedirect, redirect.Code)
	require.Equal(t, "https://example.com/a?b=1", redirect.Header().Get("Location"))
	repos := []*postgres.Repository{postgres.NewRepository(first), postgres.NewRepository(second)}
	// Claims persist across connections and expired workers cannot overwrite new claims.
	now := time.Now().UTC()
	job, err := repos[0].ClaimMetadata(ctx, now)
	require.NoError(t, err)
	_, err = repos[1].ClaimMetadata(ctx, now.Add(time.Second))
	require.ErrorIs(t, err, app.ErrMetadataUnavailable)
	reclaimed, err := repos[1].ClaimMetadata(ctx, now.Add(31*time.Second))
	require.NoError(t, err)
	require.Equal(t, int64(2), reclaimed.Attempt)
	require.NoError(t, repos[0].CompleteMetadata(ctx, job, app.Metadata{Title: "stale", FetchedAt: now}))
	_, err = repos[1].GetMetadata(ctx, job.LinkID)
	require.ErrorIs(t, err, app.ErrMetadataUnavailable)
	require.NoError(t, repos[1].CompleteMetadata(ctx, reclaimed, app.Metadata{Title: "PostgreSQL preview", FetchedAt: now}))
	preview := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, output.ShortURL, nil)
	request.Header.Set("User-Agent", "Twitterbot")
	restarted.HTTPHandler().ServeHTTP(preview, request)
	require.Equal(t, http.StatusOK, preview.Code)
	require.Contains(t, preview.Body.String(), "PostgreSQL preview")

	services := []*app.Service{app.NewService(repos[0], domain.RandomCodeGenerator{}), app.NewService(repos[1], domain.RandomCodeGenerator{})}
	userID := uuid.New()
	for _, sameKey := range []bool{true, false} {
		var wg sync.WaitGroup
		outcomes := make(chan error, 20)
		ids := make(chan uuid.UUID, 20)
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				input := app.CreateInput{LongURL: "https://example.com", IdempotencyKey: "concurrent"}
				if !sameKey {
					input.UserID = userID
					input.Plan = domain.Free
					input.IdempotencyKey = fmt.Sprintf("quota-%d", i)
				}
				output, err := services[i%2].Create(ctx, input)
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
				require.False(t, sameKey)
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
	// Force a collision to verify the adapter maps it to the application retry signal.
	candidate, err := domain.NewShortURL(uuid.New(), uuid.Nil, "https://example.com", "Ab123456", "collision-1", now, time.Hour)
	require.NoError(t, err)
	_, err = repos[0].Create(ctx, candidate, 0)
	require.NoError(t, err)
	collision, err := domain.NewShortURL(uuid.New(), uuid.Nil, "https://example.com", "Ab123456", "collision-2", now, time.Hour)
	require.NoError(t, err)
	_, err = repos[1].Create(ctx, collision, 0)
	require.ErrorIs(t, err, app.ErrCodeConflict)
	// Expired links stop resolving and cleanup cascades to their metadata jobs.
	_, err = first.Exec(ctx, "UPDATE shorten_url.links SET expires_at=$1 WHERE short_code=$2", now.Add(-366*24*time.Hour).UnixNano(), strings.TrimPrefix(output.ShortURL, "/r/"))
	require.NoError(t, err)
	expired := httptest.NewRecorder()
	restarted.HTTPHandler().ServeHTTP(expired, httptest.NewRequest(http.MethodGet, output.ShortURL, nil))
	require.Equal(t, http.StatusNotFound, expired.Code)
	require.NoError(t, services[0].Cleanup(ctx, now))
	var count int
	require.NoError(t, first.QueryRow(ctx, "SELECT count(*) FROM shorten_url.metadata WHERE link_id=$1", job.LinkID.String()).Scan(&count))
	require.Zero(t, count)
	// Concurrent metadata workers must claim each pending job at most once.
	var wg sync.WaitGroup
	jobs := make(chan app.MetadataJob, 12)
	errs := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			claimed, err := repos[i%2].ClaimMetadata(ctx, now.Add(time.Minute))
			if err == nil {
				jobs <- claimed
			} else {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(jobs)
	close(errs)
	for err := range errs {
		require.ErrorIs(t, err, app.ErrMetadataUnavailable)
	}
	seen := map[uuid.UUID]bool{}
	for claimed := range jobs {
		require.False(t, seen[claimed.LinkID], "job claimed more than once")
		seen[claimed.LinkID] = true
	}
	require.Len(t, seen, 7)
}
