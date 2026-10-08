package module

import (
	"context"

	"htqrcode/common"
	transport "htqrcode/common/grpc"
	"htqrcode/common/module/contracts"
)

type Name string

const (
	QRCode Name = "qrcode"
	Client Name = "client"
)

type Module interface {
	Name() Name
	Init(ctx context.Context) error
	RegisterHttp(ctx context.Context, e common.EchoRouter) error
	RegisterSSE(ctx context.Context, e common.EchoRouter) error
	RegisterContracts(ctx context.Context, contracts *contracts.Contracts) error
}

// GRPCModule is an optional inbound transport capability.
type GRPCModule interface {
	RegisterGRPC(context.Context, *transport.Server) error
}
