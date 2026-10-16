package app

import (
	"context"
	"time"

	"htqrcode/common/log"
)

type (
	ExpirationRepository interface {
		DeleteExpired(context.Context, time.Time) (int64, error)
	}
	// Reconcile must check live membership even when a prior eviction failed after deletion.
	ConnectionReconciler interface{ Reconcile(context.Context) error }
	Cleaner              struct {
		repo        ExpirationRepository
		now         func() time.Time
		connections ConnectionReconciler
	}
)

func NewCleaner(repo ExpirationRepository, now func() time.Time, connections ConnectionReconciler) *Cleaner {
	return &Cleaner{repo, now, connections}
}

func (c *Cleaner) Cleanup(ctx context.Context) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	count, err := c.repo.DeleteExpired(ctx, c.now().UTC())
	if err != nil {
		return count, err
	}
	if c.connections != nil {
		err = c.connections.Reconcile(ctx)
	}
	return count, err
}

func (c *Cleaner) Run(ctx context.Context) error {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	return c.RunTicks(ctx, ticker.C)
}

// RunTicks permits deterministic scheduling in lifecycle tests.
func (c *Cleaner) RunTicks(ctx context.Context, ticks <-chan time.Time) error {
	for {
		if ctx.Err() != nil {
			return nil
		}
		start := time.Now()
		count, err := c.Cleanup(ctx)
		if ctx.Err() != nil {
			return nil
		}
		logger := log.FromContext(ctx).With("duration", time.Since(start), "deleted_rooms", count)
		if err != nil {
			logger.Error("Caro cleanup failed", "error", err)
		} else {
			logger.Info("Caro cleanup complete")
		}
		select {
		case <-ctx.Done():
			return nil
		case _, ok := <-ticks:
			if !ok {
				return nil
			}
		}
	}
}
