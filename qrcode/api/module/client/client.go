package client

import "context"

type QRCode interface {
	Generate(context.Context, string) ([]byte, error)
}
