package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"

	"htqrcode/common"
)

func TestSSESkipsDeadlineAndBodyCapture(t *testing.T) {
	e := NewEcho()
	e.GET("/sse/events", func(c *echo.Context) error {
		_, deadline := c.Request().Context().Deadline()
		require.False(t, deadline)
		_, captured := c.Response().(*bodyCapturingWriter)
		require.False(t, captured)
		return c.NoContent(http.StatusOK)
	})
	e.GET("/api", func(c *echo.Context) error {
		_, deadline := c.Request().Context().Deadline()
		require.True(t, deadline)
		_, captured := c.Response().(*bodyCapturingWriter)
		require.True(t, captured)
		return c.NoContent(http.StatusOK)
	})
	for _, path := range []string{"/sse/events", "/api"} {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		require.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestSSEErrorsRemainSafeJSON(t *testing.T) {
	for _, test := range []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{"structured", common.NewInvalidInputError("invalid-request", "request is invalid").WithInternalError(errors.New("secret")), 400, `{"message":"request is invalid","slug":"invalid-request","details":[]}`},
		{"unknown", errors.New("secret"), 500, `{"message":"Internal Server Error","slug":"internal_server_error","details":[]}`},
		{"unauthorized", echo.NewHTTPError(401, "secret"), 401, `{"message":"Unauthorized","slug":"unauthorized","details":[]}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			e := NewEcho()
			e.GET("/sse/events", func(*echo.Context) error { return test.err })
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/sse/events", nil))
			require.Equal(t, test.status, rec.Code)
			require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
			require.JSONEq(t, test.body, rec.Body.String())
		})
	}
}

func TestRecoveredPanicsUseSharedContract(t *testing.T) {
	for _, path := range []string{"/api/panic", "/sse/panic"} {
		t.Run(path, func(t *testing.T) {
			e := NewEcho()
			e.GET(path, func(*echo.Context) error { panic("secret implementation detail") })
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			require.Equal(t, http.StatusInternalServerError, rec.Code)
			require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
			require.JSONEq(t, `{"message":"Internal Server Error","slug":"internal_server_error","details":[]}`, rec.Body.String())
		})
	}
}

func TestSameOriginRequests(t *testing.T) {
	for _, test := range []struct {
		name   string
		target string
		origin string
		status int
	}{
		{"no origin", "http://example.com/api", "", http.StatusOK},
		{"same origin", "http://example.com/api", "http://example.com", http.StatusOK},
		{"https", "https://example.com/api", "https://example.com", http.StatusOK},
		{"default http port", "http://example.com:80/api", "http://example.com", http.StatusOK},
		{"default https port", "https://example.com:443/api", "https://example.com", http.StatusOK},
		{"local development", "http://localhost:8080/api", "http://localhost:8080", http.StatusOK},
		{"different host", "http://example.com/api", "http://other.com", http.StatusForbidden},
		{"subdomain", "http://example.com/api", "http://sub.example.com", http.StatusForbidden},
		{"host suffix", "http://example.com/api", "http://example.com.evil.com", http.StatusForbidden},
		{"different scheme", "https://example.com/api", "http://example.com", http.StatusForbidden},
		{"different port", "http://localhost:8080/api", "http://localhost:4200", http.StatusForbidden},
		{"null origin", "http://example.com/api", "null", http.StatusForbidden},
		{"malformed origin", "http://example.com/api", "http://example.com/path", http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, method := range []string{http.MethodPost, http.MethodOptions} {
				t.Run(method, func(t *testing.T) {
					e := NewEcho()
					called := false
					e.Add(method, "/api", func(c *echo.Context) error {
						called = true
						return c.NoContent(http.StatusOK)
					})
					req := httptest.NewRequest(method, test.target, nil)
					req.Header.Set(echo.HeaderOrigin, test.origin)
					preflight := method == http.MethodOptions && test.origin != ""
					if preflight {
						req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodPost)
					}
					rec := httptest.NewRecorder()
					e.ServeHTTP(rec, req)
					status := test.status
					if preflight && status == http.StatusOK {
						status = http.StatusNoContent
					}
					require.Equal(t, status, rec.Code)
					require.Equal(t, test.status == http.StatusOK && !preflight, called)
					allowedOrigin := ""
					if test.status == http.StatusOK {
						allowedOrigin = test.origin
					}
					require.Equal(t, allowedOrigin, rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
				})
			}
		})
	}
}
