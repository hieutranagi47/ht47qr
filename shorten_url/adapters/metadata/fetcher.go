package metadata

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html"
	"htqrcode/shorten_url/app"
)

const maxHTMLBytes = 1 << 20

// Fetcher only reads bounded HTML from public HTTP(S) destinations. It never
// downloads the referenced image or executes scripts.
type Fetcher struct{ client *http.Client }

func NewFetcher() *Fetcher {
	transport := &http.Transport{
		Proxy:                  nil, // Environment proxies must not bypass address validation.
		DialContext:            publicDialContext,
		DisableKeepAlives:      true,
		TLSHandshakeTimeout:    3 * time.Second,
		ResponseHeaderTimeout:  5 * time.Second,
		MaxResponseHeaderBytes: 64 << 10,
	}
	return &Fetcher{client: &http.Client{
		Transport: transport, Timeout: 8 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			return validateURL(req.URL)
		},
	}}
}

func (f *Fetcher) Fetch(ctx context.Context, rawURL string) (app.Metadata, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return app.Metadata{}, err
	}
	if err := validateURL(target); err != nil {
		return app.Metadata{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return app.Metadata{}, err
	}
	req.Header.Set("User-Agent", "ShortURL-Metadata/1.0")
	req.Header.Set("Accept", "text/html, application/xhtml+xml")
	res, err := f.client.Do(req)
	if err != nil {
		return app.Metadata{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return app.Metadata{}, fmt.Errorf("metadata HTTP status %d", res.StatusCode)
	}
	contentType, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil || (contentType != "text/html" && contentType != "application/xhtml+xml") {
		return app.Metadata{}, errors.New("metadata response is not HTML")
	}
	if res.ContentLength > maxHTMLBytes {
		return app.Metadata{}, errors.New("metadata HTML too large")
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, maxHTMLBytes+1))
	if err != nil {
		return app.Metadata{}, err
	}
	if len(body) > maxHTMLBytes {
		return app.Metadata{}, errors.New("metadata HTML too large")
	}
	return parseHTML(strings.NewReader(string(body)), res.Request.URL), nil
}

func validateURL(u *url.URL) error {
	if u == nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
		return errors.New("metadata URL must be HTTP(S) without credentials")
	}
	if port := u.Port(); port != "" && port != "80" && port != "443" {
		return errors.New("metadata port not allowed")
	}
	if ip, err := netip.ParseAddr(u.Hostname()); err == nil && !publicIP(ip) {
		return errors.New("metadata address is not public")
	}
	return nil
}

// Block private and special-use ranges, including cloud metadata, IPv4-mapped
// IPv6, documentation networks and IPv6 transition mechanisms.
var blockedPrefixes = func() []netip.Prefix {
	var result []netip.Prefix
	for _, value := range []string{
		"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12",
		"192.0.0.0/24", "192.0.2.0/24", "192.88.99.0/24", "192.168.0.0/16", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "224.0.0.0/3",
		"2001::/23", "2001:db8::/32", "2002::/16", "3fff::/20",
	} {
		result = append(result, netip.MustParsePrefix(value))
	}
	return result
}()

func publicIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	// Only ordinary globally allocated IPv6 addresses are eligible.
	if ip.Is6() && !netip.MustParsePrefix("2000::/3").Contains(ip) {
		return false
	}
	for _, prefix := range blockedPrefixes {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

func publicDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	if port != "80" && port != "443" {
		return nil, errors.New("metadata port not allowed")
	}
	ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, errors.New("no metadata addresses")
	}
	for _, ip := range ips {
		if !publicIP(ip) {
			return nil, errors.New("metadata DNS returned a non-public address")
		}
	}
	dialer := net.Dialer{Timeout: 3 * time.Second}
	var lastErr error
	for _, ip := range ips {
		// Dial the validated IP directly; do not resolve the hostname again.
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func parseHTML(reader io.Reader, pageURL *url.URL) app.Metadata {
	tokenizer := html.NewTokenizer(reader)
	values := map[string]string{}
	var title strings.Builder
	inTitle := false
	baseURL := pageURL
	baseSeen := false
	for {
		kind := tokenizer.Next()
		if kind == html.ErrorToken {
			break
		}
		token := tokenizer.Token()
		switch kind {
		case html.StartTagToken, html.SelfClosingTagToken:
			if token.Data == "body" {
				goto done
			}
			if token.Data == "title" {
				inTitle = true
			}
			attrs := map[string]string{}
			for _, attr := range token.Attr {
				attrs[strings.ToLower(attr.Key)] = attr.Val
			}
			if token.Data == "base" && !baseSeen {
				if reference, err := url.Parse(attrs["href"]); err == nil && attrs["href"] != "" {
					candidate := pageURL.ResolveReference(reference)
					if candidate.Scheme == "http" || candidate.Scheme == "https" {
						baseURL = candidate
						baseSeen = true
					}
				}
			}
			if token.Data == "meta" {
				key := strings.ToLower(strings.TrimSpace(attrs["property"]))
				if key == "" {
					key = strings.ToLower(strings.TrimSpace(attrs["name"]))
				}
				if values[key] == "" {
					values[key] = strings.TrimSpace(attrs["content"])
				}
			}
		case html.EndTagToken:
			if token.Data == "head" {
				goto done
			}
			if token.Data == "title" {
				inTitle = false
			}
		case html.TextToken:
			if inTitle {
				title.WriteString(token.Data)
			}
		}
	}
done:
	snapshot := app.Metadata{
		Title:       bounded(first(values["og:title"], values["twitter:title"], title.String()), 512),
		Description: bounded(first(values["og:description"], values["twitter:description"], values["description"]), 2048),
	}
	image := first(values["og:image:secure_url"], values["og:image"], values["twitter:image"], values["twitter:image:src"])
	if image != "" && len(image) <= 8192 {
		if reference, err := url.Parse(image); err == nil {
			absolute := baseURL.ResolveReference(reference)
			if absolute.Hostname() != "" && absolute.User == nil && (absolute.Scheme == "http" || absolute.Scheme == "https") {
				snapshot.ImageURL = absolute.String()
			}
		}
	}
	return snapshot
}

func first(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func bounded(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= limit {
		return value
	}
	value = value[:limit]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}
