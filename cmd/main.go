package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		panic("POSTGRES_URL environment variable is not set")
	}

	dbPgx, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer dbPgx.Close()

	if err := waitForDatabase(ctx, dbPgx); err != nil {
		panic(err)
	}

	svc, err := htqrcode.New(ctx, htqrcode.ExternalServices{})
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

func waitForDatabase(ctx context.Context, db *pgxpool.Pool) error {
	startupCtx, cancel := context.WithTimeout(ctx, databaseStartupTimeout)
	defer cancel()

	ticker := time.NewTicker(databaseRetryInterval)
	defer ticker.Stop()

	var lastErr error
	for {
		pingCtx, cancel := context.WithTimeout(startupCtx, databasePingTimeout)
		lastErr = db.Ping(pingCtx)
		cancel()

		if lastErr == nil {
			return nil
		}

		select {
		case <-startupCtx.Done():
			return fmt.Errorf("PostgreSQL was not ready within %s: %w", databaseStartupTimeout, lastErr)
		case <-ticker.C:
		}
	}
}
