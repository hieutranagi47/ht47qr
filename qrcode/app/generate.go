package app

import (
	"context"
	"image"

	"htqrcode/qrcode/domain"
)

type Options struct {
	Width          uint8
	Border         int
	Circle, Custom bool
	Foreground     string
	Logo           image.Image
	Halftone       []byte
}
type Renderer interface {
	Render(context.Context, string, Options) ([]byte, error)
}
type Generator struct{ renderer Renderer }

func NewGenerator(r Renderer) Generator { return Generator{renderer: r} }
func (g Generator) Generate(ctx context.Context, message domain.MessageRequest, options Options) ([]byte, error) {
	payload, err := message.Payload()
	if err != nil {
		return nil, err
	}
	return g.renderer.Render(ctx, payload, options)
}
