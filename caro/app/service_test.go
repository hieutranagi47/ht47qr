package app

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"
	"htqrcode/caro/domain"
)

type createRepo struct {
	Repository
	calls int
	err   error
}

func (r *createRepo) Create(_ context.Context, c Creation) (Creation, error) {
	r.calls++
	return c, r.err
}

func (r *createRepo) FindCreation(context.Context, string, string, time.Time) (Creation, error) {
	return Creation{}, ErrNotFound
}

func TestCollisionBudgetAndValidation(t *testing.T) {
	r := &createRepo{err: ErrCodeConflict}
	s := NewService(r, time.Now, func() (string, error) { return "ABC234", nil })
	_, err := s.Create(context.Background(), CreateInput{IdempotencyKey: uuid.NewString(), DisplayName: "Alice", RoomType: "PUBLIC", MaxGames: 3, OwnerIP: "192.0.2.1"})
	require.ErrorIs(t, err, ErrCollisionExhausted)
	require.Equal(t, 8, r.calls)
	r.calls = 0
	_, err = s.Create(context.Background(), CreateInput{IdempotencyKey: uuid.NewString(), RoomType: "PUBLIC", MaxGames: 3, OwnerIP: "192.0.2.1"})
	require.ErrorIs(t, err, domain.ErrInvalid)
	require.Zero(t, r.calls)
	r.err = nil
	_, err = s.Create(context.Background(), CreateInput{IdempotencyKey: uuid.NewString(), DisplayName: "Alice", RoomType: "PRIVATE", MaxGames: 7, OwnerIP: "192.0.2.1"})
	require.NoError(t, err)
}

type cleanupRepo struct {
	calls   atomic.Int32
	err     error
	deleted int64
}

func (r *cleanupRepo) DeleteExpired(ctx context.Context, _ time.Time) (int64, error) {
	r.calls.Add(1)
	return r.deleted, r.err
}

type reconciler struct {
	calls atomic.Int32
	err   error
}

func (r *reconciler) Reconcile(context.Context) error { r.calls.Add(1); return r.err }
func TestCleanerSchedulingAndCancellation(t *testing.T) {
	r := &cleanupRepo{}
	n := &reconciler{}
	c := NewCleaner(r, time.Now, n)
	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time)
	done := make(chan error, 1)
	go func() { done <- c.RunTicks(ctx, ticks) }()
	require.Eventually(t, func() bool { return n.calls.Load() == 1 }, time.Second, time.Millisecond)
	ticks <- time.Now()
	require.Eventually(t, func() bool { return n.calls.Load() == 2 }, time.Second, time.Millisecond)
	cancel()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
	count, err := c.Cleanup(ctx)
	require.ErrorIs(t, err, context.Canceled)
	require.Zero(t, count)
	require.EqualValues(t, 2, r.calls.Load())
	r.err = errors.New("offline")
	_, err = c.Cleanup(context.Background())
	require.Error(t, err)
	require.EqualValues(t, 2, n.calls.Load())
}

func TestCleanerRetriesReconciliationAfterDeletion(t *testing.T) {
	repo := &cleanupRepo{}
	notify := &reconciler{err: errors.New("temporarily unavailable")}
	cleaner := NewCleaner(repo, time.Now, notify)
	_, err := cleaner.Cleanup(context.Background())
	require.Error(t, err)
	notify.err = nil
	_, err = cleaner.Cleanup(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, 2, notify.calls.Load())
	require.EqualValues(t, 2, repo.calls.Load())
}
