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
