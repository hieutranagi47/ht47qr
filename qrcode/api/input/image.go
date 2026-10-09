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
		return nil, fmt.Errorf("Image exceeds the maximum file size.")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "png" && format != "jpeg") {
		return nil, fmt.Errorf("Only valid PNG and JPEG images are allowed.")
	}
	if config.Width > maxDimension || config.Height > maxDimension {
		return nil, fmt.Errorf("Image resolution must not exceed %dx%d pixels.", maxDimension, maxDimension)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Unable to decode the image.")
	}
	return img, nil
}
