package sse

import (
	"github.com/labstack/echo/v5"

	"htqrcode/common"
	"htqrcode/common/sse"
)

type Handler struct{}

var _ ServerInterface = Handler{}

// Register adds the QR module heartbeat stream.
func Register(router common.EchoRouter) {
	RegisterHandlers(router, Handler{})
}

func (Handler) GetHeartbeat(c *echo.Context) error {
	return sse.Stream(c, nil)
}
