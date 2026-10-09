package qrcode

import (
	"context"

	"htqrcode/common"
	transport "htqrcode/common/grpc"
	"htqrcode/common/module"
	"htqrcode/common/module/contracts"
	"htqrcode/qrcode/adapters"
	grpcAPI "htqrcode/qrcode/api/grpc"
	httpAPI "htqrcode/qrcode/api/http"
	moduleAPI "htqrcode/qrcode/api/module"
	"htqrcode/qrcode/api/sse"
	"htqrcode/qrcode/app"
)

type Module struct{ generator app.Generator }

func NewModule() *Module                   { return &Module{generator: app.NewGenerator(adapters.Renderer{})} }
func (*Module) Name() module.Name          { return module.QRCode }
func (*Module) Init(context.Context) error { return nil }
func (m *Module) RegisterContracts(_ context.Context, c *contracts.Contracts) error {
	c.QRCode = moduleAPI.New(m.generator)
	return nil
}

func (m *Module) RegisterHttp(_ context.Context, router common.EchoRouter) error {
	httpAPI.Register(router, m.generator, "/api/qrcode/v1")
	return nil
}

func (*Module) RegisterSSE(_ context.Context, router common.EchoRouter) error {
	sse.Register(router)
	return nil
}

func (m *Module) RegisterGRPC(_ context.Context, server *transport.Server) error {
	grpcAPI.Register(server, m.generator)
	return nil
}

var _ module.GRPCModule = (*Module)(nil)
