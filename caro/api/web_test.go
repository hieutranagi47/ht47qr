package api

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

func TestWebRoutesAndEmbeddedBuild(t *testing.T) {
	e := echo.New()
	RegisterWeb(e)
	for _, path := range []string{"/caro/boards", "/caro/boards/", "/caro/boards/ABC234"} {
		for _, method := range []string{"GET", "HEAD"} {
			w := httptest.NewRecorder()
			e.ServeHTTP(w, httptest.NewRequest(method, path, nil))
			require.Equal(t, 200, w.Code, path)
			require.Contains(t, w.Header().Get("Content-Security-Policy"), "'wasm-unsafe-eval'")
			if method == "GET" {
				require.Contains(t, w.Body.String(), "Open boards")
			}
		}
	}
	for _, name := range []string{"app.js", "bridge.js", "style.css", "socket.io.min.js", "wasm_exec.js", "caro.wasm"} {
		w := httptest.NewRecorder()
		e.ServeHTTP(w, httptest.NewRequest("GET", "/caro/wasm/"+name, nil))
		require.Equal(t, 200, w.Code, name)
		require.NotEmpty(t, w.Body.Bytes())
		if name == "caro.wasm" {
			require.Equal(t, "application/wasm", w.Header().Get("Content-Type"))
			require.Equal(t, []byte{0, 97, 115, 109}, w.Body.Bytes()[:4])
		}
	}
	for _, path := range []string{"/caro/wasm/missing.js", "/caro/wasm/", "/caro/wasm/index.html", "/caro/wasm/nested/app.js"} {
		w := httptest.NewRecorder()
		e.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		require.Equal(t, 404, w.Code, path)
	}
}
