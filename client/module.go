package client

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"

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
	err := fs.WalkDir(staticFiles, ".", func(name string, entry fs.DirEntry, err error) error {
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
	if err != nil {
		return err
	}

	index := csp(echo.StaticFileHandler("index.html", staticFiles))
	fallback := func(c *echo.Context) error {
		// The wildcard is relative to the router, including when mounted in a group.
		requestPath := "/" + c.Param("*")
		if requestPath == "/api" || strings.HasPrefix(requestPath, "/api/") ||
			requestPath == "/assets" || strings.HasPrefix(requestPath, "/assets/") ||
			requestPath == "/media" || strings.HasPrefix(requestPath, "/media/") ||
			path.Ext(requestPath) != "" {
			return echo.NewHTTPError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}
		// Serve the app shell without redirecting or changing the browser's URL.
		return index(c)
	}
	router.GET("/*", fallback)
	router.HEAD("/*", fallback)
	return nil
}

func (*Module) RegisterSSE(context.Context, common.EchoRouter) error { return nil }

var _ module.Module = (*Module)(nil)
