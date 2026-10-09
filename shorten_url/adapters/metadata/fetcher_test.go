package metadata

import (
	"context"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseMetadata(t *testing.T) {
	page, _ := url.Parse("https://example.com/posts/a")
	snapshot := parseHTML(strings.NewReader(`<html><head><title>Fallback</title>
 <meta name="description" content="Fallback description">
 <meta property="og:title" content="OG &amp; title">
 <meta property="og:description" content="  Some   description ">
 <meta property="og:image" content="../images/card.jpg">
 </head><body><meta property="og:title" content="Ignore body"></body></html>`), page)
	require.Equal(t, "OG & title", snapshot.Title)
	require.Equal(t, "Some description", snapshot.Description)
	require.Equal(t, "https://example.com/images/card.jpg", snapshot.ImageURL)
	snapshot = parseHTML(strings.NewReader(`<head><base href="https://cdn.example.com/assets/"><title>A &amp; B</title><meta name="description" content="Desc"><meta name="twitter:image" content="card.png"></head>`), page)
	require.Equal(t, "A & B", snapshot.Title)
	require.Equal(t, "Desc", snapshot.Description)
	require.Equal(t, "https://cdn.example.com/assets/card.png", snapshot.ImageURL)
	for _, image := range []string{"javascript:alert(1)", "data:image/png;base64,abc", "https://user:secret@example.com/image"} {
		snapshot = parseHTML(strings.NewReader(`<head><meta property="og:image" content="`+image+`"></head>`), page)
		require.Empty(t, snapshot.ImageURL)
	}
	snapshot = parseHTML(strings.NewReader(`<title>`+strings.Repeat("界", 300)+`</title>`), page)
	require.LessOrEqual(t, len(snapshot.Title), 512)
	require.NotContains(t, snapshot.Title, "�")
}

func TestRejectNonPublicDestinations(t *testing.T) {
	for _, raw := range []string{
		"http://127.0.0.1/", "http://10.1.2.3/", "http://169.254.169.254/latest/meta-data",
		"http://100.100.100.200/", "http://[::1]/", "http://[::ffff:127.0.0.1]/",
		"http://[64:ff9b::a00:1]/", "http://[2002:7f00:1::]/", "http://192.0.2.1/",
		"https://user:password@example.com/", "ftp://example.com/", "http://example.com:8080/",
	} {
		t.Run(raw, func(t *testing.T) {
			target, err := url.Parse(raw)
			require.NoError(t, err)
			require.Error(t, validateURL(target))
			_, err = NewFetcher().Fetch(context.Background(), raw)
			require.Error(t, err)
		})
	}
	// Name resolution must not allow localhost even though it isn't an IP literal.
	conn, err := publicDialContext(context.Background(), "tcp", "localhost:80")
	require.Nil(t, conn)
	require.ErrorContains(t, err, "non-public")
	for _, raw := range []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111"} {
		require.True(t, publicIP(netip.MustParseAddr(raw)))
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFetcherLimitsAndRedirects(t *testing.T) {
	for _, test := range []struct {
		name, contentType, body, location string
		status                            int
		wantError                         bool
	}{
		{name: "html", contentType: "text/html; charset=utf-8", body: `<title>Title</title><meta property="og:image" content="/image.png">`, status: 200},
		{name: "not html", contentType: "image/png", body: "image bytes", status: 200, wantError: true},
		{name: "oversized", contentType: "text/html", body: strings.Repeat("a", maxHTMLBytes+1), status: 200, wantError: true},
		{name: "error response", contentType: "text/html", status: 500, wantError: true},
		{name: "private redirect", location: "http://127.0.0.1/", status: 302, wantError: true},
		{name: "redirect loop", location: "https://example.com/", status: 302, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			fetcher := NewFetcher()
			fetcher.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				require.Equal(t, "https://example.com/", req.URL.String())
				require.Empty(t, req.Header.Get("Authorization"))
				headers := http.Header{"Content-Type": []string{test.contentType}}
				if test.location != "" {
					headers.Set("Location", test.location)
				}
				return &http.Response{StatusCode: test.status, Header: headers, Body: io.NopCloser(strings.NewReader(test.body)), ContentLength: -1, Request: req}, nil
			})
			snapshot, err := fetcher.Fetch(context.Background(), "https://example.com/")
			if test.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, "Title", snapshot.Title)
				require.Equal(t, "https://example.com/image.png", snapshot.ImageURL)
			}
			if test.name == "private redirect" || test.name == "html" {
				require.Equal(t, 1, calls)
			}
		})
	}
}
