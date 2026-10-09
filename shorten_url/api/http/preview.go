package http

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/url"
	"strings"

	"github.com/labstack/echo/v5"
	"htqrcode/shorten_url/app"
)

//go:embed preview.html
var previewHTML string
var previewTemplate = template.Must(template.New("preview").Parse(previewHTML))

// User-Agent matching is for presentation only; it is not authentication.
// Unknown clients (including search crawlers) receive the normal redirect.
func isSocialPreviewBot(userAgent string) bool {
	agent := strings.ToLower(userAgent)
	// Search crawlers take precedence even if another token appears in the UA.
	for _, search := range []string{"googlebot", "bingbot", "duckduckbot", "baiduspider", "yandexbot", "applebot"} {
		if strings.Contains(agent, search) {
			return false
		}
	}
	for _, bot := range []string{
		"facebookexternalhit", "facebot", "twitterbot", "linkedinbot",
		"slackbot-linkexpanding", "discordbot", "telegrambot", "whatsapp",
		"skypeuripreview", "pinterestbot",
	} {
		if strings.Contains(agent, bot) {
			return true
		}
	}
	return false
}

func renderPreview(snapshot app.Metadata, target string) ([]byte, error) {
	data := struct{ Title, Description, ImageURL, URL string }{
		Title: snapshot.Title, Description: snapshot.Description, ImageURL: snapshot.ImageURL, URL: target,
	}
	if image, err := url.Parse(data.ImageURL); err != nil || image.Hostname() == "" || image.User != nil || (image.Scheme != "http" && image.Scheme != "https") {
		data.ImageURL = ""
	}
	if data.Title == "" {
		data.Title = target
	}
	var buffer bytes.Buffer
	err := previewTemplate.Execute(&buffer, data)
	return buffer.Bytes(), err
}

func previewCacheMiddleware(next StrictHandlerFunc, operation string) StrictHandlerFunc {
	if operation != "ResolveShortenURL" {
		return next
	}
	return func(ctx *echo.Context, request any) (any, error) {
		ctx.Response().Header().Add("Vary", "User-Agent")
		ctx.Response().Header().Set("Cache-Control", "no-store")
		return next(ctx, request)
	}
}
