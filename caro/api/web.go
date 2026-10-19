package api

import (
	"crypto/hmac"
	"crypto/sha1"
	"embed"
	"encoding/base64"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"htqrcode/common"
)

//go:embed wasm/*
var browserFiles embed.FS

// RegisterWeb serves the independent WASM client without involving the SPA assets.
func RegisterWeb(router common.EchoRouter) {
	files, err := fs.Sub(browserFiles, "wasm")
	if err != nil {
		panic(err)
	}
	secure := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Response().Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; connect-src 'self'; img-src 'self' data: blob:; object-src 'none'; base-uri 'self'; frame-ancestors 'self'")
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("Referrer-Policy", "same-origin")
			return next(c)
		}
	}
	shell := secure(echo.StaticFileHandler("index.html", files))
	// Provider credentials take precedence over the local coturn shared secret.
	// Read these at runtime so deployments never need credentials baked into assets.
	turnURLs, turnSecret := os.Getenv("CARO_TURN_URLS"), os.Getenv("CARO_TURN_SECRET")
	turnUsername, turnPassword := os.Getenv("CARO_TURN_USERNAME"), os.Getenv("CARO_TURN_PASSWORD")
	staticCredentials := turnUsername != "" || turnPassword != ""
	// ExpressTURN's free tier uses the supplied username/password directly.
	// Explicit URLs override the free-tier default, including in deployments.
	if strings.TrimSpace(turnURLs) == "" && staticCredentials {
		turnURLs = "turn:free.expressturn.com:3478?transport=udp,turn:free.expressturn.com:3478?transport=tcp"
	}
	policy := os.Getenv("CARO_ICE_TRANSPORT_POLICY")
	router.GET("/caro/rtc-config.js", secure(func(c *echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-store")
		urls := []string{}
		for _, value := range strings.Split(turnURLs, ",") {
			if value = strings.TrimSpace(value); value != "" {
				urls = append(urls, value)
			}
		}
		if len(urls) == 0 || (staticCredentials && (turnUsername == "" || turnPassword == "")) || (!staticCredentials && turnSecret == "") {
			return c.Blob(http.StatusOK, "application/javascript", []byte("/* Default WebRTC configuration. */"))
		}
		username, credential := turnUsername, turnPassword
		if !staticCredentials {
			username = strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10) + ":" + uuid.NewString()
			mac := hmac.New(sha1.New, []byte(turnSecret))
			_, _ = mac.Write([]byte(username))
			credential = base64.StdEncoding.EncodeToString(mac.Sum(nil))
		}
		config := map[string]any{"iceServers": []any{map[string]any{
			"urls": urls, "username": username, "credential": credential,
		}}}
		if policy == "relay" {
			config["iceTransportPolicy"] = policy
		}
		data, err := json.Marshal(config)
		if err != nil {
			return err
		}
		return c.Blob(http.StatusOK, "application/javascript", append(append([]byte("window.caroRTCConfig = "), data...), ';'))
	}))
	for _, route := range []string{"/caro/boards", "/caro/boards/", "/caro/boards/:room"} {
		router.GET(route, shell)
		router.HEAD(route, shell)
	}
	assets := secure(func(c *echo.Context) error {
		name := c.Param("*")
		if name == "" || strings.ContainsAny(name, "/\\") || name == "index.html" {
			return echo.NewHTTPError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}
		if strings.HasSuffix(name, ".wasm") {
			c.Response().Header().Set("Content-Type", "application/wasm")
		}
		return echo.StaticFileHandler(name, files)(c)
	})
	router.GET("/caro/wasm/*", assets)
	router.HEAD("/caro/wasm/*", assets)
}
