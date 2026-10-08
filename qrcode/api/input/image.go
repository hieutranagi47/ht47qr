// Package input contains validation shared by the QR code transports.
package input

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
)

func DecodeImage(data []byte, limit int, maxDimension int) (image.Image, error) {
	if len(data) > limit {
		return nil, fmt.Errorf("image exceeds maximum file size")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "png" && format != "jpeg") {
		return nil, fmt.Errorf("only valid PNG and JPEG images are allowed")
	}
	if config.Width > maxDimension || config.Height > maxDimension {
		return nil, fmt.Errorf("image resolution exceeds %dx%d", maxDimension, maxDimension)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("cannot decode image")
	}
	return img, nil
}
