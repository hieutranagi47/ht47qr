package caro

import (
	"context"
	"database/sql"
	"embed"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
	"htqrcode/caro/adapters/db"
	"htqrcode/caro/adapters/postgres"
	"htqrcode/caro/api"
	"htqrcode/caro/api/http"
	"htqrcode/caro/api/socketio"
	"htqrcode/caro/app"
	"htqrcode/caro/domain"
	"htqrcode/common"
	"htqrcode/common/module"
	"htqrcode/common/module/contracts"
)

type Module struct {
	database                *sql.DB
	postgres, migrationPool *pgxpool.Pool
	handler                 http.Handler
	signaling               *socketio.Server
	cleaner                 *app.Cleaner
}

func NewModule(database *sql.DB) *Module { return &Module{database: database} }
func NewPostgresModule(pool, migrationPool *pgxpool.Pool) *Module {
	if migrationPool == nil {
		migrationPool = pool
	}
	return &Module{postgres: pool, migrationPool: migrationPool}
}
func (m *Module) Name() module.Name { return module.Caro }

//go:embed adapters/db/migrations/*.sql
var migrations embed.FS

//go:embed adapters/postgres/migrations/*.sql
var postgresMigrations embed.FS

func (m *Module) Init(ctx context.Context) error {
	var repo app.Repository
	if m.postgres != nil {
		if err := common.MigratePostgresDatabaseUp(ctx, string(m.Name()), m.migrationPool, postgresMigrations, "adapters/postgres/migrations"); err != nil {
			return err
		}
		repo = postgres.NewRepository(m.postgres)
	} else {
		if err := common.MigrateSQLiteDatabaseUp(ctx, string(m.Name()), m.database, migrations, "adapters/db/migrations"); err != nil {
			return err
		}
		repo = db.NewRepository(m.database)
	}
	service := app.NewService(repo, time.Now, domain.RandomCode)
	m.handler = http.NewHandler(service)
	m.signaling = socketio.NewServer(service)
	m.cleaner = app.NewCleaner(repo, time.Now, m.signaling)
	return nil
}
func (m *Module) RegisterContracts(context.Context, *contracts.Contracts) error { return nil }
func (m *Module) RegisterHttp(ctx context.Context, e common.EchoRouter) error {
	if err := http.Register(ctx, e, m.handler); err != nil {
		return err
	}
	m.signaling.Register(e)
	api.RegisterWeb(e)
	return nil
}
func (m *Module) RegisterSSE(context.Context, common.EchoRouter) error { return nil }
func (m *Module) RunBackground(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return m.signaling.Run(ctx) })
	g.Go(func() error { return m.cleaner.Run(ctx) })
	return g.Wait()
}

var (
	_ module.Module           = (*Module)(nil)
	_ module.BackgroundModule = (*Module)(nil)
)
