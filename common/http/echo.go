package http

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"

	"htqrcode/common"
)

func NewEcho() *echo.Echo {
	e := common.NewEcho(common.EchoConfig{
		Logger:           slog.Default(),
		HTTPErrorHandler: common.EchoErrorHandler,
	})

	useMiddlewares(e)

	e.GET("/health", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	return e
}
