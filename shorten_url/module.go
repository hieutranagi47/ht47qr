package shorten_url

import (
	"context"
	"database/sql"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"

	"htqrcode/common"
	"htqrcode/common/module"
	"htqrcode/common/module/contracts"
	"htqrcode/shorten_url/adapters/db"
	"htqrcode/shorten_url/adapters/metadata"
	"htqrcode/shorten_url/adapters/postgres"
	"htqrcode/shorten_url/api/http"
	shortenURLModule "htqrcode/shorten_url/api/module"
	"htqrcode/shorten_url/app"
	"htqrcode/shorten_url/domain"
)

type Module struct {
	database      *sql.DB
	postgres      *pgxpool.Pool
	migrationPool *pgxpool.Pool
	handler       http.Handler
	metadata      *app.MetadataService
}

func NewModule(database *sql.DB) *Module { return &Module{database: database} }
func NewPostgresModule(pool, migrationPool *pgxpool.Pool) *Module {
	if migrationPool == nil {
		migrationPool = pool
	}
	return &Module{postgres: pool, migrationPool: migrationPool}
}

func (m *Module) Name() module.Name { return module.ShortenURL }

//go:embed adapters/postgres/migrations/*.sql
var postgresMigrations embed.FS

//go:embed adapters/db/migrations/*.sql
var migrations embed.FS

func (m *Module) Init(ctx context.Context) error {
	var repo interface {
		app.Repository
		app.MetadataRepository
	}
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
	service := app.NewService(repo, domain.RandomCodeGenerator{})
	m.metadata = app.NewMetadataService(repo, metadata.NewFetcher())
	m.handler = http.NewHandler(service, m.metadata)
	return nil
}

func (m *Module) RegisterContracts(_ context.Context, registry *contracts.Contracts) error {
	registry.ShortenURL = shortenURLModule.ShortenURL{}
	return nil
}

func (m *Module) RegisterHttp(ctx context.Context, e common.EchoRouter) error {
	return http.Register(ctx, e, m.handler)
}

func (m *Module) RegisterSSE(ctx context.Context, e common.EchoRouter) error {
	return nil
}

func (m *Module) RunBackground(ctx context.Context) error { return m.metadata.Run(ctx) }
