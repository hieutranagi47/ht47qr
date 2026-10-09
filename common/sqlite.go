package common

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"time"

	backoff "github.com/cenkalti/backoff/v5"
	"github.com/mattn/go-sqlite3"
)

// UpdateInSQLiteTx takes the SQLite writer lock before reading so concurrent
// read-modify-write operations preserve quotas and idempotency even across pools.
func UpdateInSQLiteTx(ctx context.Context, db *sql.DB, fn func(context.Context, *sql.Conn) error) error {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = time.Millisecond
	b.MaxInterval = 100 * time.Millisecond
	_, err := backoff.Retry(ctx, func() (struct{}, error) {
		err := updateInSQLiteTx(ctx, db, fn)
		var sqliteErr sqlite3.Error
		if err != nil && !(errors.As(err, &sqliteErr) && (sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked)) {
			return struct{}{}, backoff.Permanent(err)
		}
		return struct{}{}, err
	}, backoff.WithBackOff(b), backoff.WithMaxTries(10))
	return err
}

func updateInSQLiteTx(ctx context.Context, db *sql.DB, fn func(context.Context, *sql.Conn) error) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err = conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			rollbackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err := conn.ExecContext(rollbackCtx, "ROLLBACK"); err != nil {
				// Discard a connection whose transaction state cannot be reset.
				_ = conn.Raw(func(any) error { return driver.ErrBadConn })
			}
		}
	}()
	if err = fn(ctx, conn); err != nil {
		return err
	}
	if _, err = conn.ExecContext(ctx, "COMMIT"); err != nil {
		return err
	}
	committed = true
	return nil
}
