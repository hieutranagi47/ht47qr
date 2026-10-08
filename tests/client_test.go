package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"htqrcode"
)

func TestClientStaticFiles(t *testing.T) {
	service, err := htqrcode.New(context.Background(), htqrcode.ExternalServices{})
	require.NoError(t, err)

	for _, tc := range []struct {
		method, path string
		status       int
		body         string
	}{
		{http.MethodGet, "/", http.StatusOK, "<title>QR Code Client</title>"},
		{http.MethodGet, "/index.html", http.StatusOK, "<title>QR Code Client</title>"},
		{http.MethodHead, "/", http.StatusOK, ""},
		{http.MethodGet, "/health", http.StatusOK, "OK"},
		{http.MethodGet, "/api/v1/very-first-api", http.StatusOK, "My Handler v1"},
		{http.MethodGet, "/missing.js", http.StatusNotFound, ""},
		{http.MethodDelete, "/missing", http.StatusNotFound, ""},
		{http.MethodGet, "/module.go", http.StatusNotFound, ""},
		{http.MethodGet, "/go.mod", http.StatusNotFound, ""},
		{http.MethodGet, "/api/index.html", http.StatusNotFound, ""},
		{http.MethodGet, "/../module.go", http.StatusNotFound, ""},
		{http.MethodGet, "/%2e%2e/module.go", http.StatusNotFound, ""},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			service.HTTPHandler().ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, nil))
			require.Equal(t, tc.status, response.Code)
			if tc.body != "" {
				require.Contains(t, response.Body.String(), tc.body)
			}
			if tc.method == http.MethodHead {
				require.Empty(t, response.Body.String())
			}
			if tc.status == http.StatusOK && (strings.HasSuffix(tc.path, ".html") || tc.path == "/") {
				require.Contains(t, response.Header().Get("Content-Type"), "text/html")
			}
		})
	}

	response := httptest.NewRecorder()
	service.SSEHandler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusNotFound, response.Code)
}
