package http

import (
	"testing"

	"github.com/stretchr/testify/require"
	"htqrcode/shorten_url/app"
)

func TestSocialPreviewBotDetection(t *testing.T) {
	for _, agent := range []string{"facebookexternalhit/1.1", "Facebot", "Twitterbot/1.0", "LinkedInBot/1.0", "Slackbot-LinkExpanding 1.0", "Discordbot/2.0", "TelegramBot", "WhatsApp/2.0", "SkypeUriPreview", "Pinterestbot"} {
		require.True(t, isSocialPreviewBot(agent), agent)
	}
	for _, agent := range []string{"", "Mozilla/5.0 Chrome/100", "Googlebot/2.1", "bingbot", "DuckDuckBot", "SomeUnknownBot", "Slack-ImgProxy", "Googlebot Twitterbot"} {
		require.False(t, isSocialPreviewBot(agent), agent)
	}
}

func TestPreviewEscapesUntrustedMetadata(t *testing.T) {
	page, err := renderPreview(app.Metadata{Title: `</title><script>alert(1)</script>`, Description: `" onload="alert(1)`, ImageURL: `javascript:alert(1)`}, "https://example.com/?a=1&b=2")
	require.NoError(t, err)
	require.NotContains(t, string(page), "<script>")
	require.Contains(t, string(page), "&lt;script&gt;")
	require.Contains(t, string(page), "&#34;")
	require.NotContains(t, string(page), "javascript:")
	require.Contains(t, string(page), `name="twitter:card" content="summary"`)
	require.Contains(t, string(page), "<body></body>")
	require.Contains(t, string(page), "https://example.com/?a=1&amp;b=2")
}
