package adapters

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"htqrcode/qrcode/app"
)

type (
	Renderer struct{}
	buffer   struct{ bytes.Buffer }
)

func (*buffer) Close() error { return nil }
func (Renderer) Render(ctx context.Context, payload string, o app.Options) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	qr, err := qrcode.NewWith(payload, qrcode.WithEncodingMode(qrcode.EncModeByte), qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionHighest))
	if err != nil {
		return nil, err
	}
	opts := []standard.ImageOption{standard.WithBuiltinImageEncoder(standard.PNG_FORMAT), standard.WithLogoSizeMultiplier(2), standard.WithQRWidth(o.Width), standard.WithBorderWidth(o.Border)}
	if o.Foreground != "" {
		opts = append(opts, standard.WithFgColorRGBHex(o.Foreground))
	}
	if o.Logo != nil {
		opts = append(opts, standard.WithLogoImage(o.Logo))
	}
	if o.Circle {
		opts = append(opts, standard.WithCircleShape())
	} else {
		if o.Custom {
			opts = append(opts, standard.WithCustomShape(newShape(0.7)))
		}
		if len(o.Halftone) != 0 {
			f, err := os.CreateTemp("", "qrcode-halftone-*.img")
			if err != nil {
				return nil, err
			}
			defer os.Remove(f.Name())
			_, writeErr := f.Write(o.Halftone)
			closeErr := f.Close()
			if writeErr != nil {
				return nil, writeErr
			}
			if closeErr != nil {
				return nil, closeErr
			}
			opts = append(opts, standard.WithHalftone(f.Name()))
		}
	}
	var output buffer
	w := standard.NewWithWriter(&output, opts...)
	// Bound raster allocation even for large payloads combined with large blocks.
	attr := w.Attribute(qr.Dimension())
	if attr.W > 4096 || attr.H > 4096 {
		return nil, fmt.Errorf("generated image exceeds 4096 pixels")
	}
	if err := qr.Save(w); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

type smallerCircle struct{ smallerPercent float64 }

func (sc *smallerCircle) DrawFinder(ctx *standard.DrawContext) {
	backup := sc.smallerPercent
	sc.smallerPercent = 1.0
	sc.Draw(ctx)
	sc.smallerPercent = backup
}

func newShape(radiusPercent float64) standard.IShape {
	return &smallerCircle{smallerPercent: radiusPercent}
}

func (sc *smallerCircle) Draw(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	x, y := ctx.UpperLeft()
	color := ctx.Color()

	// choose a proper radius values
	radius := w / 2
	r2 := h / 2
	if r2 <= radius {
		radius = r2
	}

	// Scale the circle to the requested fraction of the block radius.
	radius = int(float64(radius) * sc.smallerPercent)

	cx, cy := x+float64(w)/2.0, y+float64(h)/2.0 // get center point
	ctx.DrawCircle(cx, cy, float64(radius))
	ctx.SetColor(color)
	ctx.Fill()
}
