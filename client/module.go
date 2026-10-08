package client

import (
	"context"
	"embed"
	"io/fs"
	"path"

	"github.com/labstack/echo/v5"

	"htqrcode/common"
	"htqrcode/common/module"
	"htqrcode/common/module/contracts"
)

// Only the static API directory is included in the binary.
//
//go:embed api
var files embed.FS

type Module struct{}

func NewModule() *Module                   { return &Module{} }
func (*Module) Name() module.Name          { return module.Client }
func (*Module) Init(context.Context) error { return nil }
func (*Module) RegisterContracts(context.Context, *contracts.Contracts) error {
	return nil
}

func (*Module) RegisterHttp(_ context.Context, router common.EchoRouter) error {
	staticFiles := echo.MustSubFS(files, "api")
	return fs.WalkDir(staticFiles, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		handler := echo.StaticFileHandler(name, staticFiles)
		router.GET("/"+name, handler)
		router.HEAD("/"+name, handler)
		if path.Base(name) == "index.html" {
			url := path.Dir("/"+name) + "/"
			if url == "//" {
				url = "/"
			}
			router.GET(url, handler)
			router.HEAD(url, handler)
		}
		return nil
	})
}

func (*Module) RegisterSSE(context.Context, common.EchoRouter) error { return nil }

var _ module.Module = (*Module)(nil)
