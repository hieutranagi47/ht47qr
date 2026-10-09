// Package module provides shorten_url's public API for other modules.
package module

import "htqrcode/shorten_url/domain"

// ShortenURL exposes services provided by the shorten_url module.
type ShortenURL struct{}

// GenerateCode creates an eight-character, cryptographically random Base62 code.
func (ShortenURL) GenerateCode() (string, error) {
	return domain.RandomCodeGenerator{}.Generate()
}
