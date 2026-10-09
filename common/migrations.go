package common

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"

	migrate "github.com/golang-migrate/migrate/v4"
	sqliteMigrate "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"

	"htqrcode/common/log"
)

func MigrateDatabaseUp(
	ctx context.Context,
	moduleName string,
	databaseURL string,
	fs fs.FS,
	migrationsDir string,
) error {
	db, err := sql.Open("sqlite3", databaseURL)
	if err != nil {
		return fmt.Errorf("could not open SQLite migration database: %w", err)
	}
	defer db.Close()

	return MigrateSQLiteDatabaseUp(ctx, moduleName, db, fs, migrationsDir)
}

// MigrateSQLiteDatabaseUp applies module migrations without closing the caller's database.
func MigrateSQLiteDatabaseUp(ctx context.Context, moduleName string, db *sql.DB, fs fs.FS, migrationsDir string) error {
	d, err := iofs.New(fs, migrationsDir)
	if err != nil {
		return fmt.Errorf("could not create iofs driver: %w", err)
	}

	defer d.Close()

	migDb, err := sqliteMigrate.WithInstance(db, &sqliteMigrate.Config{
		DatabaseName:    string(moduleName),
		MigrationsTable: string(moduleName) + "_schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("could not connect to SQLite migrations database: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "sqlite3", migDb)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	finished := make(chan struct{})
	defer close(finished)

	go func() {
		select {
		case <-finished:
			return
		case <-ctx.Done():
			log.FromContext(ctx).Info("Interrupt received, stopping migrations...")
			m.GracefulStop <- true
		}
	}()

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration up failed: %w", err)
	}

	return nil
}
