package api

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"htqrcode/common"
)

//go:embed wasm/*
var browserFiles embed.FS

// RegisterWeb serves the independent WASM client without involving the SPA assets.
func RegisterWeb(router common.EchoRouter) {
	files, err := fs.Sub(browserFiles, "wasm")
	if err != nil {
		panic(err)
	}
	secure := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Response().Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; connect-src 'self'; img-src 'self' data: blob:; object-src 'none'; base-uri 'self'; frame-ancestors 'self'")
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("Referrer-Policy", "same-origin")
			return next(c)
		}
	}
	shell := secure(echo.StaticFileHandler("index.html", files))
	for _, route := range []string{"/caro/boards", "/caro/boards/", "/caro/boards/:room"} {
		router.GET(route, shell)
		router.HEAD(route, shell)
	}
	assets := secure(func(c *echo.Context) error {
		name := c.Param("*")
		if name == "" || strings.ContainsAny(name, "/\\") || name == "index.html" {
			return echo.NewHTTPError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}
		if strings.HasSuffix(name, ".wasm") {
			c.Response().Header().Set("Content-Type", "application/wasm")
		}
		return echo.StaticFileHandler(name, files)(c)
	})
	router.GET("/caro/wasm/*", assets)
	router.HEAD("/caro/wasm/*", assets)
}
