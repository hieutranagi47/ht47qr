package client

import (
	"context"
	"embed"
	"io/fs"
	"path"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"htqrcode/common"
	"htqrcode/common/module"
	"htqrcode/common/module/contracts"
)

// Only the static API directory is included in the binary.
//
//go:embed api
var files embed.FS

// Match the frontend's CSP, including inline scripts/styles and QR image previews.
const contentSecurityPolicy = "connect-src 'self'; " +
	"font-src 'self'; frame-src 'self'; img-src 'self' data: blob:; " +
	"manifest-src 'self'; media-src 'self'; object-src 'none'; " +
	"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
	"script-src-elem 'self' 'unsafe-eval' 'unsafe-inline'; " +
	"style-src 'self' 'unsafe-inline' 'unsafe-hashes'; " +
	"worker-src 'self'; child-src 'none'; base-uri 'self';"

type Module struct{}

func NewModule() *Module                   { return &Module{} }
func (*Module) Name() module.Name          { return module.Client }
func (*Module) Init(context.Context) error { return nil }
func (*Module) RegisterContracts(context.Context, *contracts.Contracts) error {
	return nil
}

func (*Module) RegisterHttp(_ context.Context, router common.EchoRouter) error {
	staticFiles := echo.MustSubFS(files, "api")
	csp := middleware.SecureWithConfig(middleware.SecureConfig{
		ContentSecurityPolicy: contentSecurityPolicy,
	})
	return fs.WalkDir(staticFiles, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		handler := echo.StaticFileHandler(name, staticFiles)
		router.GET("/"+name, handler, csp)
		router.HEAD("/"+name, handler, csp)
		if path.Base(name) == "index.html" {
			url := path.Dir("/"+name) + "/"
			if url == "//" {
				url = "/"
			}
			router.GET(url, handler, csp)
			router.HEAD(url, handler, csp)
		}
		return nil
	})
}

func (*Module) RegisterSSE(context.Context, common.EchoRouter) error { return nil }

var _ module.Module = (*Module)(nil)
