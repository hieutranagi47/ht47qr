package adapters_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"htqrcode/caro"
	"htqrcode/caro/adapters/db"
	"htqrcode/caro/adapters/postgres"
	"htqrcode/caro/app"
	"htqrcode/caro/domain"
)

func testRepositories(t *testing.T, fn func(*testing.T, app.Repository)) {
	t.Helper()
	t.Run("sqlite", func(t *testing.T) {
		pool, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "caro.db")+"?_busy_timeout=5000")
		require.NoError(t, err)
		t.Cleanup(func() { pool.Close() })
		m := caro.NewModule(pool)
		require.NoError(t, m.Init(context.Background()))
		require.NoError(t, m.Init(context.Background()))
		fn(t, db.NewRepository(pool))
	})
	t.Run("postgres", func(t *testing.T) {
		url := os.Getenv("CARO_TEST_POSTGRES_URL")
		if url == "" {
			t.Skip("CARO_TEST_POSTGRES_URL not configured; use a disposable database")
		}
		pool, err := pgxpool.New(context.Background(), url)
		require.NoError(t, err)
		t.Cleanup(pool.Close)
		lock, err := pool.Acquire(context.Background())
		require.NoError(t, err)
		_, err = lock.Exec(context.Background(), "SELECT pg_advisory_lock(47470001)")
		require.NoError(t, err)
		t.Cleanup(func() { _, _ = lock.Exec(context.Background(), "SELECT pg_advisory_unlock(47470001)"); lock.Release() })
		m := caro.NewPostgresModule(pool, nil)
		require.NoError(t, m.Init(context.Background()))
		require.NoError(t, m.Init(context.Background()))
		_, err = pool.Exec(context.Background(), "TRUNCATE caro.rooms CASCADE")
		require.NoError(t, err)
		t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "TRUNCATE caro.rooms CASCADE") })
		fn(t, postgres.NewRepository(pool))
	})
}

func TestRoomTransactions(t *testing.T) {
	testRepositories(t, func(t *testing.T, repo app.Repository) {
		ctx := context.Background()
		now := time.Now().UTC().Truncate(time.Microsecond)
		s := app.NewService(repo, func() time.Time { return now }, domain.RandomCode)
		created, err := s.Create(ctx, app.CreateInput{IdempotencyKey: uuid.NewString(), DisplayName: "Alice", RoomType: "PUBLIC", MaxGames: 3, OwnerIP: "192.0.2.1"})
		require.NoError(t, err)
		code := created.Room.Code()
		rooms, err := s.List(ctx)
		require.NoError(t, err)
		require.Len(t, rooms, 1)
		stored, err := s.Get(ctx, code)
		require.NoError(t, err)
		require.Equal(t, now.Add(24*time.Hour), stored.ExpiresAt())
		require.Equal(t, domain.HashCredential(created.HostToken), stored.Players()[0].CredentialHash())
		require.ErrorIs(t, persistRoom(ctx, repo, created.Room), app.ErrCodeConflict)
		// A mutation returning an error must roll back all state.
		rollback := errors.New("rollback")
		_, err = repo.Update(ctx, code, now, func(r *domain.Room) error {
			_, _, e := r.Join("temp", "Bob", "temp", "", now, "198.51.100.99")
			require.NoError(t, e)
			return rollback
		})
		require.ErrorIs(t, err, rollback)
		stored, err = s.Get(ctx, code)
		require.NoError(t, err)
		require.Len(t, stored.Players(), 1)
		// A reserved host is never claimable without its credential.
		_, err = s.Join(ctx, code, "", "host", "wrong", "198.51.100.99")
		require.ErrorIs(t, err, domain.ErrUnauthorized)
		joined, err := s.Join(ctx, code, "", "host", created.HostToken, "198.51.100.99")
		require.NoError(t, err)
		require.Equal(t, domain.Waiting, joined.Room.Status())
		var wg sync.WaitGroup
		var success atomic.Int32
		failures := make(chan error, 12)
		for i := range 12 {
			wg.Go(func() {
				_, err := s.Join(ctx, code, "Guest", string(rune('a'+i)), "", "198.51.100.99")
				if err == nil {
					success.Add(1)
				} else {
					failures <- err
				}
			})
		}
		wg.Wait()
		close(failures)
		require.EqualValues(t, 1, success.Load())
		for err := range failures {
			require.ErrorIs(t, err, domain.ErrFull)
		}
		stored, err = s.Get(ctx, code)
		require.NoError(t, err)
		require.Equal(t, domain.Playing, stored.Status())
		require.Len(t, stored.Players(), 2)
		duplicate, err := s.Join(ctx, code, "", "host", created.HostToken, "198.51.100.99")
		require.NoError(t, err)
		require.False(t, duplicate.Changed)
		peer, err := s.Peer(ctx, code, "host")
		require.NoError(t, err)
		require.NotEmpty(t, peer)
		_, err = s.Peer(ctx, code, "outsider")
		require.ErrorIs(t, err, domain.ErrUnauthorized)
		rooms, err = s.List(ctx)
		require.NoError(t, err)
		require.Empty(t, rooms)
		_, err = s.Leave(ctx, code, peer)
		require.NoError(t, err)
		stored, err = s.Get(ctx, code)
		require.NoError(t, err)
		require.Equal(t, domain.Finished, stored.Status())
		_, err = s.Join(ctx, code, "New", "new", "", "198.51.100.99")
		require.ErrorIs(t, err, domain.ErrUnavailable)
		now = now.Add(24 * time.Hour)
		_, err = s.Get(ctx, code)
		require.ErrorIs(t, err, app.ErrNotFound)
		_, err = s.Join(ctx, code, "New", "new", "", "198.51.100.99")
		require.ErrorIs(t, err, domain.ErrUnavailable)
		n, err := repo.DeleteExpired(ctx, now.Add(-time.Nanosecond))
		require.NoError(t, err)
		require.Zero(t, n)
		n, err = repo.DeleteExpired(ctx, now)
		require.NoError(t, err)
		require.EqualValues(t, 1, n)
		n, err = repo.DeleteExpired(ctx, now)
		require.NoError(t, err)
		require.Zero(t, n)
	})
}

func TestVisibilityAndAtomicHostReservation(t *testing.T) {
	testRepositories(t, func(t *testing.T, repo app.Repository) {
		ctx := context.Background()
		now := time.Now().UTC().Truncate(time.Microsecond)
		s := app.NewService(repo, func() time.Time { return now }, domain.RandomCode)
		private, err := s.Create(ctx, app.CreateInput{IdempotencyKey: uuid.NewString(), DisplayName: "Alice", RoomType: "PRIVATE", MaxGames: 1, OwnerIP: "192.0.2.1"})
		require.NoError(t, err)
		public, err := s.Create(ctx, app.CreateInput{IdempotencyKey: uuid.NewString(), DisplayName: "Alice", RoomType: "PUBLIC", MaxGames: 7, OwnerIP: "192.0.2.2"})
		require.NoError(t, err)
		list, err := s.List(ctx)
		require.NoError(t, err)
		require.Len(t, list, 1)
		require.Equal(t, public.Room.ID(), list[0].ID())
		// Force a player primary-key failure after inserting the room; neither insert may survive.
		candidate, err := domain.RestoreRoom("rollback-room", "ZZZ234", "PUBLIC", domain.Waiting, 3, now, now.Add(domain.Lifetime), private.Room.Players(), "192.0.2.3")
		require.NoError(t, err)
		require.Error(t, persistRoom(ctx, repo, candidate))
		_, err = repo.FindActive(ctx, candidate.Code(), now)
		require.ErrorIs(t, err, app.ErrNotFound)
		// A connected socket cannot be admitted to another room.
		_, err = s.Join(ctx, private.Room.Code(), "", "shared", private.HostToken, "198.51.100.99")
		require.NoError(t, err)
		_, err = s.Join(ctx, public.Room.Code(), "Bob", "shared", "", "198.51.100.99")
		require.ErrorIs(t, err, app.ErrMembershipConflict)
		stored, err := s.Get(ctx, public.Room.Code())
		require.NoError(t, err)
		require.Len(t, stored.Players(), 1)
		require.NoError(t, s.ResetConnections(ctx))
		stored, err = s.Get(ctx, private.Room.Code())
		require.NoError(t, err)
		require.Equal(t, domain.Finished, stored.Status())
		require.Empty(t, stored.Players()[0].Socket())
		// Cascades must free unique player IDs even when the DSN did not enable foreign keys.
		count, err := repo.DeleteExpired(ctx, now.Add(domain.Lifetime))
		require.NoError(t, err)
		require.EqualValues(t, 2, count)
		require.NoError(t, persistRoom(ctx, repo, private.Room))
		count, err = repo.DeleteExpired(ctx, now.Add(domain.Lifetime))
		require.NoError(t, err)
		require.EqualValues(t, 1, count)
	})
}

func TestExpiredJoinAndCleanupRace(t *testing.T) {
	testRepositories(t, func(t *testing.T, repo app.Repository) {
		ctx := context.Background()
		clock := &atomic.Int64{}
		clock.Store(time.Now().Truncate(time.Microsecond).UnixNano())
		s := app.NewService(repo, func() time.Time { return time.Unix(0, clock.Load()) }, domain.RandomCode)
		created, err := s.Create(ctx, app.CreateInput{IdempotencyKey: uuid.NewString(), DisplayName: "Host", RoomType: "PUBLIC", MaxGames: 1, OwnerIP: "192.0.2.1"})
		require.NoError(t, err)
		clock.Add(int64(domain.Lifetime))
		var wg sync.WaitGroup
		failures := make(chan error, 8)
		for i := range 8 {
			wg.Go(func() {
				_, err := s.Join(ctx, created.Room.Code(), "Guest", fmt.Sprintf("expired-%d", i), "", "198.51.100.99")
				failures <- err
			})
		}
		count, err := repo.DeleteExpired(ctx, time.Unix(0, clock.Load()))
		require.NoError(t, err)
		require.EqualValues(t, 1, count)
		wg.Wait()
		close(failures)
		for err := range failures {
			require.True(t, errors.Is(err, app.ErrNotFound) || errors.Is(err, domain.ErrUnavailable), "%v", err)
		}
		_, err = s.Get(ctx, created.Room.Code())
		require.ErrorIs(t, err, app.ErrNotFound)
	})
}

func TestOwnerRoomLimits(t *testing.T) {
	testRepositories(t, func(t *testing.T, repo app.Repository) {
		ctx := context.Background()
		now := time.Now().UTC().Truncate(time.Microsecond)
		s := app.NewService(repo, func() time.Time { return now }, domain.RandomCode)
		input := app.CreateInput{IdempotencyKey: uuid.NewString(), DisplayName: "Owner", RoomType: "PUBLIC", MaxGames: 3, OwnerIP: "::ffff:192.0.2.10"}
		active, err := s.Create(ctx, newRequest(input))
		require.NoError(t, err)
		require.Equal(t, "192.0.2.10", active.Room.OwnerIP())
		input.OwnerIP = "192.0.2.10"
		_, err = s.Create(ctx, newRequest(input))
		require.ErrorIs(t, err, domain.ErrWaitingRoomLimit)
		other := input
		other.OwnerIP = "192.0.2.11"
		_, err = s.Create(ctx, other)
		require.NoError(t, err)
		_, err = s.Join(ctx, active.Room.Code(), "", "host-1", active.HostToken, "198.51.100.99")
		require.NoError(t, err)
		_, err = s.Join(ctx, active.Room.Code(), "Guest", "guest-1", "", "198.51.100.99")
		require.NoError(t, err)
		waiting, err := s.Create(ctx, newRequest(input))
		require.NoError(t, err)
		_, err = s.Create(ctx, newRequest(input))
		require.ErrorIs(t, err, domain.ErrWaitingRoomLimit)
		_, err = s.Join(ctx, waiting.Room.Code(), "", "host-2", waiting.HostToken, "198.51.100.99")
		require.NoError(t, err)
		_, err = s.Join(ctx, waiting.Room.Code(), "Guest", "guest-2", "", "198.51.100.99")
		require.ErrorIs(t, err, domain.ErrActiveRoomLimit)
		unchanged, err := s.Get(ctx, waiting.Room.Code())
		require.NoError(t, err)
		require.Equal(t, domain.Waiting, unchanged.Status())
		require.Len(t, unchanged.Players(), 1)
		// A rejected start leaves no guest binding behind and does not affect the active room.
		activeRoom, err := s.Get(ctx, active.Room.Code())
		require.NoError(t, err)
		require.Equal(t, domain.Playing, activeRoom.Status())
		_, err = s.Leave(ctx, active.Room.Code(), "guest-1")
		require.NoError(t, err)
		_, err = s.Join(ctx, waiting.Room.Code(), "Guest", "guest-2", "", "198.51.100.99")
		require.NoError(t, err)
		_, err = s.Create(ctx, newRequest(input))
		require.NoError(t, err)
		// Expiry releases both quotas without waiting for physical deletion.
		now = now.Add(domain.Lifetime)
		_, err = s.Create(ctx, newRequest(input))
		require.NoError(t, err)
	})
}

func TestConcurrentOwnerCreationLimit(t *testing.T) {
	testRepositories(t, func(t *testing.T, repo app.Repository) {
		now := time.Now().UTC().Truncate(time.Microsecond)
		// Separate application instances must share a persistent quota, not a local mutex.
		services := []*app.Service{app.NewService(repo, func() time.Time { return now }, domain.RandomCode), app.NewService(repo, func() time.Time { return now }, domain.RandomCode)}
		var wg sync.WaitGroup
		var successes atomic.Int32
		failures := make(chan error, 12)
		for i := range 12 {
			wg.Go(func() {
				_, err := services[i%2].Create(context.Background(), app.CreateInput{IdempotencyKey: uuid.NewString(), DisplayName: "Owner", RoomType: "PRIVATE", MaxGames: 1, OwnerIP: "2001:db8::1"})
				if err == nil {
					successes.Add(1)
				} else {
					failures <- err
				}
			})
		}
		wg.Wait()
		close(failures)
		require.EqualValues(t, 1, successes.Load())
		for err := range failures {
			require.ErrorIs(t, err, domain.ErrWaitingRoomLimit)
		}
	})
}

func persistRoom(ctx context.Context, repo app.Repository, room domain.Room) error {
	nonce, err := domain.NewCredential()
	if err != nil {
		return err
	}
	_, err = repo.Create(ctx, app.Creation{Room: room, KeyHash: domain.HashCredential(uuid.NewString()), Fingerprint: "fixture", Nonce: nonce})
	return err
}

func newRequest(in app.CreateInput) app.CreateInput { in.IdempotencyKey = uuid.NewString(); return in }

func TestCreationIdempotency(t *testing.T) {
	testRepositories(t, func(t *testing.T, repo app.Repository) {
		ctx := context.Background()
		now := time.Now().UTC().Truncate(time.Microsecond)
		s := app.NewService(repo, func() time.Time { return now }, domain.RandomCode)
		input := app.CreateInput{DisplayName: "Alice", RoomType: "PUBLIC", MaxGames: 3, OwnerIP: "192.0.2.1", IdempotencyKey: uuid.NewString()}
		first, err := s.Create(ctx, input)
		require.NoError(t, err)
		retry, err := s.Create(ctx, input)
		require.NoError(t, err)
		require.Equal(t, first, retry)
		restarted := app.NewService(repo, func() time.Time { return now }, func() (string, error) { return "", errors.New("generator unavailable") })
		retry, err = restarted.Create(ctx, input)
		require.NoError(t, err)
		require.Equal(t, first, retry)
		changed := input
		changed.MaxGames = 7
		_, err = s.Create(ctx, changed)
		require.ErrorIs(t, err, app.ErrIdempotencyConflict)
		_, err = s.Join(ctx, first.Room.Code(), "", "host", first.HostToken, input.OwnerIP)
		require.NoError(t, err)
		_, err = s.Join(ctx, first.Room.Code(), "Guest", "guest", "", "192.0.2.2")
		require.NoError(t, err)
		retry, err = s.Create(ctx, input)
		require.NoError(t, err)
		require.Equal(t, first, retry)
		_, err = s.Leave(ctx, first.Room.Code(), "guest")
		require.NoError(t, err)
		retry, err = s.Create(ctx, input)
		require.NoError(t, err)
		require.Equal(t, first, retry)
		now = now.Add(domain.Lifetime)
		_, err = s.Create(ctx, input)
		require.ErrorIs(t, err, app.ErrIdempotencyExpired)
		_, err = s.Create(ctx, newRequest(input))
		require.NoError(t, err)
	})
}

func TestConcurrentCreationRetries(t *testing.T) {
	testRepositories(t, func(t *testing.T, repo app.Repository) {
		now := time.Now().UTC().Truncate(time.Microsecond)
		s := app.NewService(repo, func() time.Time { return now }, domain.RandomCode)
		input := app.CreateInput{DisplayName: "Alice", RoomType: "PUBLIC", MaxGames: 1, OwnerIP: "2001:db8::1", IdempotencyKey: uuid.NewString()}
		type outcome struct {
			result app.Created
			err    error
		}
		outcomes := make(chan outcome, 12)
		var wg sync.WaitGroup
		for range 12 {
			wg.Go(func() { created, err := s.Create(context.Background(), input); outcomes <- outcome{created, err} })
		}
		wg.Wait()
		close(outcomes)
		var first *app.Created
		for result := range outcomes {
			require.NoError(t, result.err)
			if first == nil {
				first = &result.result
			} else {
				require.Equal(t, *first, result.result)
			}
		}
		rooms, err := s.List(context.Background())
		require.NoError(t, err)
		require.Len(t, rooms, 1)
		require.Len(t, rooms[0].Players(), 1)
	})
}

func TestGuestParticipationLimits(t *testing.T) {
	testRepositories(t, func(t *testing.T, repo app.Repository) {
		ctx := context.Background()
		now := time.Now().UTC().Truncate(time.Microsecond)
		s := app.NewService(repo, func() time.Time { return now }, domain.RandomCode)
		create := func(ip string) app.Created {
			r, err := s.Create(ctx, app.CreateInput{DisplayName: "Host", RoomType: "PUBLIC", MaxGames: 1, OwnerIP: ip, IdempotencyKey: uuid.NewString()})
			require.NoError(t, err)
			return r
		}
		first := create("192.0.2.1")
		second := create("192.0.2.2")
		const guestIP = "198.51.100.10"
		_, err := s.Join(ctx, first.Room.Code(), "Guest", "guest-1", "", guestIP)
		require.NoError(t, err)
		// Waiting guest membership consumes the same quota as a created host reservation.
		_, err = s.Join(ctx, second.Room.Code(), "Guest", "guest-2", "", guestIP)
		require.ErrorIs(t, err, domain.ErrWaitingRoomLimit)
		_, err = s.Create(ctx, app.CreateInput{DisplayName: "Guest", RoomType: "PRIVATE", MaxGames: 1, OwnerIP: guestIP, IdempotencyKey: uuid.NewString()})
		require.ErrorIs(t, err, domain.ErrWaitingRoomLimit)
		_, err = s.Join(ctx, first.Room.Code(), "", "host-1", first.HostToken, "192.0.2.1")
		require.NoError(t, err)
		// One active room plus one waiting room is allowed for guests too.
		_, err = s.Join(ctx, second.Room.Code(), "Guest", "guest-2", "", guestIP)
		require.NoError(t, err)
		_, err = s.Join(ctx, second.Room.Code(), "", "host-2", second.HostToken, "192.0.2.2")
		require.ErrorIs(t, err, domain.ErrActiveRoomLimit)
		stored, err := s.Get(ctx, second.Room.Code())
		require.NoError(t, err)
		require.Equal(t, domain.Waiting, stored.Status())
		require.Empty(t, stored.Players()[0].Socket())
		_, err = s.Leave(ctx, first.Room.Code(), "guest-1")
		require.NoError(t, err)
		_, err = s.Join(ctx, second.Room.Code(), "", "host-2", second.HostToken, "192.0.2.2")
		require.NoError(t, err)
		// Guest departure releases participation; finished rooms consume neither quota.
		_, err = s.Leave(ctx, second.Room.Code(), "guest-2")
		require.NoError(t, err)
		_ = create(guestIP)
	})
}

func TestConcurrentCrossRoomGuestJoins(t *testing.T) {
	testRepositories(t, func(t *testing.T, repo app.Repository) {
		ctx := context.Background()
		now := time.Now().UTC().Truncate(time.Microsecond)
		s := app.NewService(repo, func() time.Time { return now }, domain.RandomCode)
		rooms := make([]app.Created, 2)
		for i := range rooms {
			r, err := s.Create(ctx, app.CreateInput{DisplayName: "Host", RoomType: "PUBLIC", MaxGames: 1, OwnerIP: fmt.Sprintf("192.0.2.%d", i+1), IdempotencyKey: uuid.NewString()})
			require.NoError(t, err)
			rooms[i] = r
		}
		results := make(chan error, 2)
		var wg sync.WaitGroup
		for i := range 2 {
			wg.Go(func() {
				_, err := s.Join(ctx, rooms[i].Room.Code(), "Guest", fmt.Sprintf("guest-%d", i), "", "198.51.100.1")
				results <- err
			})
		}
		wg.Wait()
		close(results)
		successes := 0
		for err := range results {
			if err == nil {
				successes++
			} else {
				require.ErrorIs(t, err, domain.ErrWaitingRoomLimit)
			}
		}
		require.Equal(t, 1, successes)
	})
}
