package common

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	migrate "github.com/golang-migrate/migrate/v4"
	pgxMigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"htqrcode/common/log"
)

func MigratePostgresDatabaseUp(
	ctx context.Context,
	moduleName string,
	pool *pgxpool.Pool,
	fs fs.FS,
	migrationsDir string,
) error {
	log.FromContext(ctx).Info("applying PostgreSQL migrations", "schema", moduleName,
		"host", pool.Config().ConnConfig.Host, "database", pool.Config().ConnConfig.Database)
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	d, err := iofs.New(fs, migrationsDir)
	if err != nil {
		return fmt.Errorf("could not create iofs driver: %w", err)
	}

	defer d.Close()

	if _, err := db.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+pgx.Identifier{moduleName}.Sanitize()); err != nil {
		return fmt.Errorf("could not create schema %s: %w", moduleName, err)
	}

	migDb, err := pgxMigrate.WithInstance(db, &pgxMigrate.Config{
		SchemaName:      string(moduleName),
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("could not connect to pgx migrations database: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "pgx", migDb)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	// Release the dedicated database connection acquired by pgxMigrate.WithInstance.
	// Without this, each MigrateDatabaseUp call leaks one pgxpool connection.
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.FromContext(ctx).With("error", srcErr).Error("closing migration source failed")
		}
		if dbErr != nil {
			log.FromContext(ctx).With("error", dbErr).Error("closing migration database failed")
		}
	}()

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

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("could not read migration version: %w", err)
	}
	log.FromContext(ctx).Info("PostgreSQL migrations complete", "schema", moduleName,
		"version", version, "dirty", dirty)
	return nil
}
