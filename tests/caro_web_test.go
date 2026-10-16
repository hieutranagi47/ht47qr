package tests

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"htqrcode"
)

func TestCaroWebComponent(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "web.db")+"?_busy_timeout=5000")
	require.NoError(t, err)
	defer db.Close()
	service, err := htqrcode.New(context.Background(), htqrcode.ExternalServices{Database: db})
	require.NoError(t, err)
	for _, path := range []string{"/caro/boards", "/caro/boards/ABC234", "/caro/wasm/app.js", "/caro/wasm/caro.wasm"} {
		w := httptest.NewRecorder()
		service.HTTPHandler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		require.Equal(t, 200, w.Code, path)
		require.Contains(t, w.Header().Get("Content-Security-Policy"), "wasm-unsafe-eval", path)
		if path == "/caro/boards" {
			require.Contains(t, w.Body.String(), "Your next move")
		}
	}
}
