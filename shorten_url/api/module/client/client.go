// Package client declares the shorten_url module API consumed by other modules.
package client

type ShortenURL interface {
	GenerateCode() (string, error)
}
