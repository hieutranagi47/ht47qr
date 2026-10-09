package shorten_url_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"htqrcode/shorten_url/adapters/db"
	"htqrcode/shorten_url/app"
	"htqrcode/shorten_url/domain"
)

type metadataFetcherFunc func(context.Context, string) (app.Metadata, error)

func (f metadataFetcherFunc) Fetch(ctx context.Context, target string) (app.Metadata, error) {
	return f(ctx, target)
}

func TestMetadataPersistenceAndPreviewRouting(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "links.db")
	database := openDatabase(t, path)
	service := newService(t, database)
	repo := db.NewRepository(database)
	linkService := app.NewService(repo, domain.RandomCodeGenerator{})
	input := app.CreateInput{LongURL: "https://example.com/article?a=1&b=2", IdempotencyKey: "preview"}
	created, err := linkService.Create(ctx, input)
	require.NoError(t, err)
	worker := app.NewMetadataService(repo, metadataFetcherFunc(func(ctx context.Context, target string) (app.Metadata, error) {
		require.Equal(t, input.LongURL, target)
		return app.Metadata{Title: `Title <script>alert(1)</script>`, Description: `A "quoted" description`, ImageURL: "https://cdn.example.com/card.jpg"}, nil
	}))
	route := "/r/" + created.URL.Code()
	request := func(handler http.Handler, agent string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, route, nil)
		req.Header.Set("User-Agent", agent)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		require.Contains(t, strings.Join(res.Header().Values("Vary"), ","), "User-Agent")
		require.Equal(t, "no-store", res.Header().Get("Cache-Control"))
		return res
	}
	require.Equal(t, 307, request(service.HTTPHandler(), "Twitterbot").Code)
	processed, err := worker.ProcessNext(ctx)
	require.NoError(t, err)
	require.True(t, processed)
	processed, err = worker.ProcessNext(ctx)
	require.NoError(t, err)
	require.False(t, processed)
	replay, err := linkService.Create(ctx, input)
	require.NoError(t, err)
	require.True(t, replay.Existing)
	var jobs int
	require.NoError(t, database.QueryRow("SELECT count(*) FROM shorten_url_metadata").Scan(&jobs))
	require.Equal(t, 1, jobs)
	require.NoError(t, database.Close())
	restartedDB := openDatabase(t, path)
	restarted := newService(t, restartedDB)
	for _, agent := range []string{"Twitterbot", "Slackbot-LinkExpanding", "facebookexternalhit", "Discordbot"} {
		res := request(restarted.HTTPHandler(), agent)
		require.Equal(t, 200, res.Code)
		require.Contains(t, res.Header().Get("Content-Type"), "text/html")
		require.Empty(t, res.Header().Get("Location"))
		require.Contains(t, res.Body.String(), `property="og:image" content="https://cdn.example.com/card.jpg"`)
		require.Contains(t, res.Body.String(), `name="twitter:card" content="summary_large_image"`)
		require.NotContains(t, res.Body.String(), "<script>")
		require.Contains(t, res.Body.String(), "<body></body>")
	}
	for _, agent := range []string{"Mozilla/5.0", "Googlebot/2.1", "bingbot", "UnknownBot", ""} {
		res := request(restarted.HTTPHandler(), agent)
		require.Equal(t, 307, res.Code)
		require.Equal(t, input.LongURL, res.Header().Get("Location"))
	}
	_, err = restartedDB.Exec("UPDATE shorten_url_links SET expires_at = ?", time.Now().Add(-time.Second).UnixNano())
	require.NoError(t, err)
	require.Equal(t, 404, request(restarted.HTTPHandler(), "Twitterbot").Code)
}

func TestMetadataJobLeasesAndRetries(t *testing.T) {
	ctx := context.Background()
	database := openDatabase(t, filepath.Join(t.TempDir(), "links.db"))
	_ = newService(t, database)
	repo := db.NewRepository(database)
	created, err := app.NewService(repo, domain.RandomCodeGenerator{}).Create(ctx, app.CreateInput{LongURL: "https://example.com", IdempotencyKey: "retry"})
	require.NoError(t, err)
	now := time.Now()
	job, err := repo.ClaimMetadata(ctx, now)
	require.NoError(t, err)
	require.Equal(t, int64(1), job.Attempt)
	_, err = repo.ClaimMetadata(ctx, now.Add(time.Second))
	require.ErrorIs(t, err, app.ErrMetadataUnavailable)
	reclaimed, err := repo.ClaimMetadata(ctx, now.Add(31*time.Second))
	require.NoError(t, err)
	require.Equal(t, int64(2), reclaimed.Attempt)
	// An expired worker cannot overwrite the current claim.
	require.NoError(t, repo.CompleteMetadata(ctx, job, app.Metadata{Title: "stale", FetchedAt: now}))
	_, err = repo.GetMetadata(ctx, created.URL.ID())
	require.ErrorIs(t, err, app.ErrMetadataUnavailable)
	require.NoError(t, repo.FailMetadata(ctx, reclaimed, now))
	var calls atomic.Int64
	worker := app.NewMetadataService(repo, metadataFetcherFunc(func(context.Context, string) (app.Metadata, error) {
		calls.Add(1)
		return app.Metadata{}, errors.New("unreachable")
	}))
	processed, err := worker.ProcessNext(ctx)
	require.NoError(t, err)
	require.True(t, processed)
	processed, err = worker.ProcessNext(ctx)
	require.NoError(t, err)
	require.False(t, processed)
	require.Equal(t, int64(1), calls.Load())
	var status string
	require.NoError(t, database.QueryRow("SELECT status FROM shorten_url_metadata").Scan(&status))
	require.Equal(t, "failed", status)
	// Cancellation ends the worker without leaking a background goroutine.
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	require.NoError(t, worker.Run(cancelled))
}

func TestEmbeddedBackgroundWorkerCancellation(t *testing.T) {
	database := openDatabase(t, filepath.Join(t.TempDir(), "links.db"))
	service := newService(t, database)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- service.RunBackground(ctx) }()
	cancel()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("background worker did not stop")
	}
}
