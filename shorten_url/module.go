package shorten_url

import (
	"context"
	"database/sql"
	"embed"

	"htqrcode/common"
	"htqrcode/common/module"
	"htqrcode/common/module/contracts"
	"htqrcode/shorten_url/adapters/db"
	"htqrcode/shorten_url/api/http"
	shortenURLModule "htqrcode/shorten_url/api/module"
	"htqrcode/shorten_url/app"
	"htqrcode/shorten_url/domain"
)

type Module struct {
	database *sql.DB
	handler  http.Handler
}

func NewModule(database *sql.DB) *Module { return &Module{database: database} }
func (m *Module) Name() module.Name      { return module.ShortenURL }

//go:embed adapters/db/migrations/*.sql
var migrations embed.FS

func (m *Module) Init(ctx context.Context) error {
	repo := db.NewRepository(m.database)
	service := app.NewService(repo, domain.RandomCodeGenerator{})
	m.handler = http.NewHandler(service)
	return common.MigrateSQLiteDatabaseUp(ctx, string(m.Name()), m.database, migrations, "adapters/db/migrations")
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
