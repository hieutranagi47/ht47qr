package client

import (
	"context"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"

	"htqrcode/common"
)

func TestClientContentSecurityPolicy(t *testing.T) {
	paths := []string{"/"}
	require.NoError(t, fs.WalkDir(echo.MustSubFS(files, "api"), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			paths = append(paths, "/"+name)
		}
		return nil
	}))

	for _, prefix := range []string{"", "/client"} {
		t.Run("prefix="+prefix, func(t *testing.T) {
			e := echo.New()
			var router common.EchoRouter = e
			if prefix != "" {
				router = e.Group(prefix)
			}
			require.NoError(t, NewModule().RegisterHttp(context.Background(), router))
			e.GET("/health", func(c *echo.Context) error {
				return c.String(http.StatusOK, "OK")
			})

			for _, method := range []string{http.MethodGet, http.MethodHead} {
				for _, path := range paths {
					t.Run(method+" "+prefix+path, func(t *testing.T) {
						response := httptest.NewRecorder()
						e.ServeHTTP(response, httptest.NewRequest(method, prefix+path, nil))
						require.Equal(t, http.StatusOK, response.Code)
						require.Equal(t, []string{contentSecurityPolicy}, response.Header().Values(echo.HeaderContentSecurityPolicy))
						require.Empty(t, response.Header().Get(echo.HeaderContentSecurityPolicyReportOnly))
						if method == http.MethodHead {
							require.Empty(t, response.Body.String())
						}
					})
				}
			}

			for _, path := range []string{"/health", prefix + "/missing.js"} {
				response := httptest.NewRecorder()
				e.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
				require.Empty(t, response.Header().Get(echo.HeaderContentSecurityPolicy))
			}
		})
	}
}

func TestClientSPAFallback(t *testing.T) {
	index, err := files.ReadFile("api/index.html")
	require.NoError(t, err)

	for _, prefix := range []string{"", "/client"} {
		t.Run("prefix="+prefix, func(t *testing.T) {
			e := echo.New()
			var router common.EchoRouter = e
			if prefix != "" {
				router = e.Group(prefix)
			}
			require.NoError(t, NewModule().RegisterHttp(context.Background(), router))
			for _, route := range []string{"/health", "/qrcode/docs", "/api/v1/example"} {
				router.GET(route, func(c *echo.Context) error {
					return c.String(http.StatusOK, "backend")
				})
				response := httptest.NewRecorder()
				e.ServeHTTP(response, httptest.NewRequest(http.MethodGet, prefix+route, nil))
				require.Equal(t, http.StatusOK, response.Code)
				require.Equal(t, "backend", response.Body.String())
				require.Empty(t, response.Header().Get(echo.HeaderContentSecurityPolicy))
			}

			for _, method := range []string{http.MethodGet, http.MethodHead} {
				for _, route := range []string{"/interview", "/qrcode", "/qrcode/history/123?tab=details", "/apiary"} {
					t.Run(method+" "+route, func(t *testing.T) {
						request := httptest.NewRequest(method, prefix+route, nil)
						originalURL := request.URL.String()
						response := httptest.NewRecorder()
						e.ServeHTTP(response, request)
						require.Equal(t, http.StatusOK, response.Code)
						require.Equal(t, contentSecurityPolicy, response.Header().Get(echo.HeaderContentSecurityPolicy))
						require.Contains(t, response.Header().Get(echo.HeaderContentType), "text/html")
						require.Empty(t, response.Header().Get(echo.HeaderLocation))
						require.Equal(t, originalURL, request.URL.String())
						if method == http.MethodHead {
							require.Empty(t, response.Body.String())
						} else {
							require.Equal(t, string(index), response.Body.String())
						}
					})
				}

				for _, route := range []string{"/api", "/api/unknown", "/missing.js", "/assets/missing", "/media/missing"} {
					response := httptest.NewRecorder()
					e.ServeHTTP(response, httptest.NewRequest(method, prefix+route, nil))
					require.Equal(t, http.StatusNotFound, response.Code, method+" "+route)
					require.Empty(t, response.Header().Get(echo.HeaderContentSecurityPolicy))
				}
			}

			for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
				response := httptest.NewRecorder()
				e.ServeHTTP(response, httptest.NewRequest(method, prefix+"/interview", nil))
				require.Equal(t, http.StatusMethodNotAllowed, response.Code)
				require.Empty(t, response.Header().Get(echo.HeaderContentSecurityPolicy))
			}
		})
	}
}
