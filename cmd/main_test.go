package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenDatabasesSQLiteFallback(t *testing.T) {
	for _, name := range []string{"DATABASE_URL", "POSTGRES_URL", "DATABASE_URL_UNPOOLED", "POSTGRES_URL_NON_POOLING"} {
		t.Setenv(name, "")
	}
	path := filepath.Join(t.TempDir(), "nested", "links.db")
	t.Setenv("SQLITE_PATH", path)
	services, cleanup, err := openDatabases(context.Background())
	require.NoError(t, err)
	defer cleanup()
	require.NotNil(t, services.Database)
	require.Nil(t, services.Postgres)
	var mode string
	require.NoError(t, services.Database.QueryRow("PRAGMA journal_mode").Scan(&mode))
	require.Equal(t, "wal", mode)
	_, err = os.Stat(path)
	require.NoError(t, err)
}

func TestOpenDatabasesPostgresSelection(t *testing.T) {
	dsn := os.Getenv("SHORTEN_URL_TEST_POSTGRES_URL")
	if dsn == "" {
		t.Skip("set SHORTEN_URL_TEST_POSTGRES_URL")
	}
	t.Setenv("DATABASE_URL", dsn)
	t.Setenv("POSTGRES_URL", "invalid fallback")
	t.Setenv("DATABASE_URL_UNPOOLED", dsn)
	t.Setenv("POSTGRES_URL_NON_POOLING", "invalid fallback")
	services, cleanup, err := openDatabases(context.Background())
	require.NoError(t, err)
	require.Nil(t, services.Database)
	require.NotNil(t, services.Postgres)
	require.NotNil(t, services.PostgresMigrations)
	require.NoError(t, services.Postgres.Ping(context.Background()))
	require.NoError(t, services.PostgresMigrations.Ping(context.Background()))
	cleanup()
	require.Error(t, services.Postgres.Ping(context.Background()))
	require.Error(t, services.PostgresMigrations.Ping(context.Background()))
	// The Vercel alias also works when the primary URL is absent.
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_URL_UNPOOLED", "")
	t.Setenv("POSTGRES_URL", dsn)
	t.Setenv("POSTGRES_URL_NON_POOLING", "")
	services, cleanup, err = openDatabases(context.Background())
	require.NoError(t, err)
	defer cleanup()
	require.NotNil(t, services.Postgres)
	require.Nil(t, services.PostgresMigrations)
}
