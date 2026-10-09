package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jackc/pgx/v5/pgxpool"
	"htqrcode"
	"htqrcode/common/log"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env file before main runs
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using system environment variables")
	}
}

const (
	databaseStartupTimeout = 30 * time.Second
	databasePingTimeout    = time.Second
	databaseRetryInterval  = time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Init(slog.LevelInfo)

	services, closeDatabases, err := openDatabases(ctx)
	if err != nil {
		panic(err)
	}
	defer closeDatabases()

	svc, err := htqrcode.New(ctx, services)
	if err != nil {
		panic(err)
	}

	httpPort := os.Getenv("SERVER_PORT")
	httpsPort := os.Getenv("SERVER_PORT_TLS")
	ssePort := os.Getenv("SERVER_SSE_PORT")
	ssesPort := os.Getenv("SERVER_SSE_PORT_TLS")

	if err := svc.Run(ctx, httpPort, httpsPort, ssePort, ssesPort, os.Getenv("SERVER_GRPC_PORT")); err != nil {
		panic(err)
	}
}

type postgresPinger struct{ *pgxpool.Pool }

func (p postgresPinger) PingContext(ctx context.Context) error { return p.Ping(ctx) }

func waitForDatabase(ctx context.Context, db interface{ PingContext(context.Context) error }) error {
	startupCtx, cancel := context.WithTimeout(ctx, databaseStartupTimeout)
	defer cancel()

	ticker := time.NewTicker(databaseRetryInterval)
	defer ticker.Stop()

	var lastErr error
	for {
		pingCtx, cancel := context.WithTimeout(startupCtx, databasePingTimeout)
		lastErr = db.PingContext(pingCtx)
		cancel()

		if lastErr == nil {
			return nil
		}

		select {
		case <-startupCtx.Done():
			return fmt.Errorf("database was not ready within %s: %w", databaseStartupTimeout, lastErr)
		case <-ticker.C:
		}
	}
}

// PostgreSQL is selected only when a connection URL is configured. SQLite
// keeps its existing file path and connection settings for local deployments.
func openDatabases(ctx context.Context) (htqrcode.ExternalServices, func(), error) {
	dsn := firstEnvironment("DATABASE_URL", "POSTGRES_URL")
	backend := strings.ToLower(strings.TrimSpace(os.Getenv("DATABASE_BACKEND")))
	switch backend {
	case "", "auto":
		// Preserve automatic selection for existing local deployments.
	case "postgres":
		if dsn == "" {
			return htqrcode.ExternalServices{}, nil, fmt.Errorf("DATABASE_BACKEND=postgres requires DATABASE_URL or POSTGRES_URL at runtime")
		}
	case "sqlite":
		dsn = ""
	default:
		return htqrcode.ExternalServices{}, nil, fmt.Errorf("DATABASE_BACKEND must be auto, postgres, or sqlite")
	}
	if dsn != "" {
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			return htqrcode.ExternalServices{}, nil, fmt.Errorf("invalid PostgreSQL configuration")
		}
		log.FromContext(ctx).Info("database selected", "backend", "postgresql",
			"host", pool.Config().ConnConfig.Host, "database", pool.Config().ConnConfig.Database)
		cleanup := func() { pool.Close() }
		if err := waitForDatabase(ctx, postgresPinger{pool}); err != nil {
			cleanup()
			return htqrcode.ExternalServices{}, nil, fmt.Errorf("PostgreSQL startup failed: %w", err)
		}
		services := htqrcode.ExternalServices{Postgres: pool}
		if migrationDSN := firstEnvironment("DATABASE_URL_UNPOOLED", "POSTGRES_URL_NON_POOLING"); migrationDSN != "" {
			migrationPool, err := pgxpool.New(ctx, migrationDSN)
			if err != nil {
				cleanup()
				return htqrcode.ExternalServices{}, nil, fmt.Errorf("invalid PostgreSQL migration configuration")
			}
			log.FromContext(ctx).Info("migration database selected", "backend", "postgresql",
				"host", migrationPool.Config().ConnConfig.Host, "database", migrationPool.Config().ConnConfig.Database)
			cleanup = func() { migrationPool.Close(); pool.Close() }
			if err := waitForDatabase(ctx, postgresPinger{migrationPool}); err != nil {
				cleanup()
				return htqrcode.ExternalServices{}, nil, fmt.Errorf("PostgreSQL migration startup failed: %w", err)
			}
			services.PostgresMigrations = migrationPool
		}
		return services, cleanup, nil
	}
	dbPath := os.Getenv("SQLITE_PATH")
	if dbPath == "" {
		dbPath = "./data/htqrcode.db"
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		return htqrcode.ExternalServices{}, nil, fmt.Errorf("could not create SQLite database directory: %w", err)
	}
	log.FromContext(ctx).Info("database selected", "backend", "sqlite", "path", dbPath,
		"configured_backend", backend)
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		return htqrcode.ExternalServices{}, nil, err
	}
	if err := waitForDatabase(ctx, db); err != nil {
		db.Close()
		return htqrcode.ExternalServices{}, nil, err
	}
	return htqrcode.ExternalServices{Database: db}, func() { db.Close() }, nil
}

func firstEnvironment(names ...string) string {
	for _, name := range names {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}
