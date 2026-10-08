package module

import (
	"context"

	"htqrcode/qrcode/app"
	"htqrcode/qrcode/domain"
)

type QRCode struct{ generator app.Generator }

func New(g app.Generator) QRCode { return QRCode{generator: g} }
func (a QRCode) Generate(ctx context.Context, text string) ([]byte, error) {
	return a.generator.Generate(ctx, domain.MessageRequest{Text: text}, app.Options{Width: 10, Border: 10})
}
