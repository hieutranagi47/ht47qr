package caro

import (
	"context"
	"database/sql"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"
	"htqrcode/caro/adapters/db"
	"htqrcode/caro/app"
	"htqrcode/caro/domain"
	"htqrcode/common"
)

func TestOwnerMigrationPreservesLegacyRooms(t *testing.T) {
	ctx := context.Background()
	pool, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "upgrade.db"))
	require.NoError(t, err)
	defer pool.Close()
	old := fstest.MapFS{}
	for _, name := range []string{"adapters/db/migrations/000001_rooms.up.sql", "adapters/db/migrations/000001_rooms.down.sql"} {
		content, err := fs.ReadFile(migrations, name)
		require.NoError(t, err)
		old[name] = &fstest.MapFile{Data: content}
	}
	require.NoError(t, common.MigrateSQLiteDatabaseUp(ctx, "caro", pool, old, "adapters/db/migrations"))
	now := time.Now().UTC()
	_, err = pool.ExecContext(ctx, `INSERT INTO caro_rooms(id,code,room_type,max_games,status,created_at,expires_at) VALUES('legacy','ABC234','PUBLIC',1,'WAITING',?,?)`, now.UnixNano(), now.Add(domain.Lifetime).UnixNano())
	require.NoError(t, err)
	_, err = pool.ExecContext(ctx, `INSERT INTO caro_room_players(id,room_id,display_name,player_role,credential_hash,joined_at) VALUES('legacy-host','legacy','Host','PLAYER_1',?,?)`, domain.HashCredential("legacy-token"), now.UnixNano())
	require.NoError(t, err)
	m := NewModule(pool)
	require.NoError(t, m.Init(ctx))
	require.NoError(t, m.Init(ctx))
	service := app.NewService(db.NewRepository(pool), func() time.Time { return now }, domain.RandomCode)
	room, err := service.Get(ctx, "ABC234")
	require.NoError(t, err)
	require.Empty(t, room.OwnerIP())
	require.Len(t, room.Players(), 1)
	_, err = service.Create(ctx, app.CreateInput{IdempotencyKey: uuid.NewString(), DisplayName: "New owner", RoomType: "PUBLIC", MaxGames: 3, OwnerIP: "192.0.2.1"})
	require.NoError(t, err)
	rooms, err := service.List(ctx)
	require.NoError(t, err)
	require.Len(t, rooms, 2)
}
