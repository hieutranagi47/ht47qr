package api

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

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

func TestRTCConfig(t *testing.T) {
	for _, configured := range []bool{false, true} {
		t.Run(strconv.FormatBool(configured), func(t *testing.T) {
			t.Setenv("CARO_TURN_URLS", "")
			t.Setenv("CARO_TURN_SECRET", "")
			t.Setenv("CARO_TURN_USERNAME", "")
			t.Setenv("CARO_TURN_PASSWORD", "")
			t.Setenv("CARO_ICE_TRANSPORT_POLICY", "relay")
			if configured {
				t.Setenv("CARO_TURN_URLS", "turn:127.0.0.1:3478?transport=tcp, turn:127.0.0.1:3478?transport=udp")
				t.Setenv("CARO_TURN_SECRET", "test-private-secret")
			}
			e := echo.New()
			RegisterWeb(e)
			request := func() string {
				w := httptest.NewRecorder()
				e.ServeHTTP(w, httptest.NewRequest("GET", "/caro/rtc-config.js", nil))
				require.Equal(t, 200, w.Code)
				require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
				require.Contains(t, w.Header().Get("Content-Type"), "application/javascript")
				return w.Body.String()
			}
			body := request()
			if !configured {
				require.NotContains(t, body, "window.caroRTCConfig")
				return
			}
			require.NotContains(t, body, "test-private-secret")
			var config struct {
				ICEServers []struct {
					URLs                 []string `json:"urls"`
					Username, Credential string
				} `json:"iceServers"`
				Policy string `json:"iceTransportPolicy"`
			}
			require.NoError(t, json.Unmarshal([]byte(strings.TrimSuffix(strings.TrimPrefix(body, "window.caroRTCConfig = "), ";")), &config))
			require.Equal(t, "relay", config.Policy)
			require.Len(t, config.ICEServers, 1)
			server := config.ICEServers[0]
			require.Equal(t, []string{"turn:127.0.0.1:3478?transport=tcp", "turn:127.0.0.1:3478?transport=udp"}, server.URLs)
			expiry, err := strconv.ParseInt(strings.SplitN(server.Username, ":", 2)[0], 10, 64)
			require.NoError(t, err)
			require.WithinDuration(t, time.Now().Add(time.Hour), time.Unix(expiry, 0), 2*time.Second)
			mac := hmac.New(sha1.New, []byte("test-private-secret"))
			_, _ = mac.Write([]byte(server.Username))
			require.Equal(t, base64.StdEncoding.EncodeToString(mac.Sum(nil)), server.Credential)
			require.NotEqual(t, body, request(), "each page load receives fresh credentials")
		})
	}
}

func TestRTCConfigProviderCredentials(t *testing.T) {
	for _, tc := range []struct {
		name, urls, username, password, policy string
		configured                             bool
	}{
		{"provider", "turn:turn.example.com:3478?transport=udp, turn:turn.example.com:3478?transport=tcp", "sample-turn-user", "sample-turn-password+/=", "all", true},
		{"relay only", "turn:turn.example.com:3478", "sample-turn-user", "sample-turn-password", "relay", true},
		{"missing password", "turn:turn.example.com:3478", "sample-turn-user", "", "all", false},
		{"missing username", "turn:turn.example.com:3478", "", "sample-turn-password", "all", false},
		{"empty URLs", " , ", "sample-turn-user", "sample-turn-password", "all", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CARO_TURN_URLS", tc.urls)
			t.Setenv("CARO_TURN_USERNAME", tc.username)
			t.Setenv("CARO_TURN_PASSWORD", tc.password)
			// Provider credentials must override a leftover local coturn secret.
			t.Setenv("CARO_TURN_SECRET", "test-private-secret")
			t.Setenv("CARO_ICE_TRANSPORT_POLICY", tc.policy)
			e := echo.New()
			RegisterWeb(e)
			w := httptest.NewRecorder()
			e.ServeHTTP(w, httptest.NewRequest("GET", "/caro/rtc-config.js", nil))
			require.Equal(t, 200, w.Code)
			require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
			body := w.Body.String()
			require.NotContains(t, body, "test-private-secret")
			if !tc.configured {
				require.NotContains(t, body, "window.caroRTCConfig")
				return
			}
			var config struct {
				ICEServers []struct {
					URLs                 []string `json:"urls"`
					Username, Credential string
				} `json:"iceServers"`
				Policy string `json:"iceTransportPolicy"`
			}
			require.NoError(t, json.Unmarshal([]byte(strings.TrimSuffix(strings.TrimPrefix(body, "window.caroRTCConfig = "), ";")), &config))
			require.Len(t, config.ICEServers, 1)
			server := config.ICEServers[0]
			require.Equal(t, tc.username, server.Username)
			require.Equal(t, tc.password, server.Credential)
			var urls []string
			for _, value := range strings.Split(tc.urls, ",") {
				urls = append(urls, strings.TrimSpace(value))
			}
			require.Equal(t, urls, server.URLs)
			if tc.policy == "relay" {
				require.Equal(t, "relay", config.Policy)
			} else {
				require.Empty(t, config.Policy, "default policy allows direct and relay connections")
			}
		})
	}
}
