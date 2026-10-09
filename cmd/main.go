package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

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

	dbPath := os.Getenv("SQLITE_PATH")
	if dbPath == "" {
		dbPath = "./data/htqrcode.db"
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		panic(fmt.Errorf("could not create SQLite database directory: %w", err))
	}
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := waitForDatabase(ctx, db); err != nil {
		panic(err)
	}

	svc, err := htqrcode.New(ctx, htqrcode.ExternalServices{Database: db})
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

func waitForDatabase(ctx context.Context, db *sql.DB) error {
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
			return fmt.Errorf("SQLite was not ready within %s: %w", databaseStartupTimeout, lastErr)
		case <-ticker.C:
		}
	}
}
