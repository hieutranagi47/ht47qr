package htqrcode

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcAPI "htqrcode/qrcode/api/grpc"

	commonHTTP "htqrcode/common/http"
	qrcodeSSE "htqrcode/qrcode/api/sse"
)

func TestServiceRunRejectsInvalidTLSBeforeStartingListeners(t *testing.T) {
	server := httptest.NewTLSServer(http.NotFoundHandler())
	server.Close()
	pair := server.TLS.Certificates[0]
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: pair.Certificate[0]})
	keyDER, err := x509.MarshalPKCS8PrivateKey(pair.PrivateKey)
	require.NoError(t, err)
	key := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	for _, test := range []struct {
		name string
		cert string
		key  string
	}{
		{name: "missing", cert: "", key: ""},
		{name: "file paths", cert: "external/cert/localhost+1.pem", key: "external/cert/localhost+1-key.pem"},
		{name: "escaped newlines", cert: strings.ReplaceAll(string(cert), "\n", `\n`), key: string(key)},
		{name: "folded newlines", cert: strings.ReplaceAll(string(cert), "\n", " "), key: string(key)},
		{name: "invalid key", cert: string(cert), key: "invalid"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("TLS_CERT", test.cert)
			t.Setenv("TLS_KEY", test.key)
			// No routers are initialized: invalid TLS must fail before listeners start.
			var service Service
			err := service.Run(context.Background(), "0", "0", "0", "0")
			require.ErrorContains(t, err, "TLS_CERT")
			require.ErrorContains(t, err, "TLS_KEY")
		})
	}
}

func TestServiceRunReportsListenerFailure(t *testing.T) {
	cert, err := os.ReadFile("external/cert/localhost+1.pem")
	if err != nil {
		t.Fatal(err)
	}
	key, err := os.ReadFile("external/cert/localhost+1-key.pem")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TLS_CERT", string(cert))
	t.Setenv("TLS_KEY", string(key))

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	svc, err := New(ctx, ExternalServices{})
	if err != nil {
		t.Fatal(err)
	}
	// An occupied port must fail promptly without waiting for a message router.
	if err := svc.Run(ctx, port, port, port, port); err == nil {
		t.Fatal("Run succeeded with an occupied port")
	}
	if ctx.Err() != nil {
		t.Fatal("Run waited for cancellation instead of starting the listeners")
	}
}

func TestSSEListenersServeAndShutdown(t *testing.T) {
	for _, secure := range []bool{false, true} {
		t.Run(map[bool]string{false: "http", true: "https"}[secure], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			router := commonHTTP.NewEcho()
			qrcodeSSE.Register(router.Group("/sse"))
			config := serverConfig(ctx, "0", true)
			// Use an explicit loopback listener so the test never exposes a service.
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			scheme := "http"
			if secure {
				scheme = "https"
				certificate, err := tls.LoadX509KeyPair("external/cert/localhost+1.pem", "external/cert/localhost+1-key.pem")
				if err != nil {
					t.Fatal(err)
				}
				listener = tls.NewListener(listener, &tls.Config{Certificates: []tls.Certificate{certificate}})
			}
			config.Listener = listener
			done := make(chan error, 1)
			go func() {
				var cert, key []byte
				if secure {
					cert, err = os.ReadFile("external/cert/localhost+1.pem")
					if err == nil {
						key, err = os.ReadFile("external/cert/localhost+1-key.pem")
					}
					if err != nil {
						done <- err
						return
					}
				}
				done <- serveHTTP(ctx, config, router, secure, cert, key)
			}()
			client := &http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
			defer client.CloseIdleConnections()
			response, err := client.Get(scheme + "://" + listener.Addr().String() + "/sse/qrcode/heartbeat")
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != 200 || response.Header.Get("Content-Type") != "text/event-stream" {
				t.Fatalf("unexpected stream response: %s %v", response.Status, response.Header)
			}
			reader := bufio.NewReader(response.Body)
			frame := make([]byte, len(": connected\n\n"))
			if _, err := io.ReadFull(reader, frame); err != nil {
				t.Fatal(err)
			}
			if string(frame) != ": connected\n\n" {
				t.Fatalf("initial frame = %q", frame)
			}
			cancel()
			select {
			case err := <-done:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("server did not shut down with an active stream")
			}
			if _, err := reader.ReadByte(); err != io.EOF {
				t.Fatalf("stream error = %v, want EOF", err)
			}
		})
	}
}

type shutdownListener struct {
	net.Listener
	closed chan struct{}
}

func (l *shutdownListener) Close() error {
	err := l.Listener.Close()
	select {
	case <-l.closed:
	default:
		close(l.closed)
	}
	return err
}

func TestHTTPShutdown(t *testing.T) {
	for _, timeout := range []bool{false, true} {
		t.Run(map[bool]string{false: "drains request", true: "force closes after deadline"}[timeout], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			ln := &shutdownListener{Listener: listener, closed: make(chan struct{})}
			defer ln.Close()
			started := make(chan context.Context, 1)
			release := make(chan struct{})
			defer close(release)
			finished := make(chan struct{})
			router := commonHTTP.NewEcho()
			router.GET("/work", func(c *echo.Context) error {
				defer close(finished)
				started <- c.Request().Context()
				select {
				case <-release:
					return c.String(http.StatusOK, "completed")
				case <-c.Request().Context().Done():
					return nil
				}
			})
			config := serverConfig(ctx, "0", false)
			config.Listener = ln
			if timeout {
				config.GracefulTimeout = 50 * time.Millisecond
			}
			done := make(chan error, 1)
			go func() { done <- serveHTTP(ctx, config, router, false, nil, nil) }()
			client := &http.Client{Timeout: 3 * time.Second}
			defer client.CloseIdleConnections()
			result := make(chan error, 1)
			go func() {
				response, err := client.Get("http://" + ln.Addr().String() + "/work")
				if err == nil {
					defer response.Body.Close()
					var body []byte
					body, err = io.ReadAll(response.Body)
					if err == nil && string(body) != "completed" {
						err = errors.New("request did not complete")
					}
				}
				result <- err
			}()
			var requestCtx context.Context
			select {
			case requestCtx = <-started:
			case <-time.After(3 * time.Second):
				t.Fatal("handler did not start")
			}
			cancel()
			select {
			case <-ln.closed:
			case <-time.After(3 * time.Second):
				t.Fatal("listener did not close")
			}
			if !timeout {
				if err := requestCtx.Err(); err != nil {
					t.Fatalf("request canceled before draining: %v", err)
				}
				release <- struct{}{}
			}
			select {
			case err := <-done:
				if timeout && !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("shutdown error = %v, want deadline exceeded", err)
				}
				if !timeout && err != nil {
					t.Fatal(err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("server did not stop")
			}
			select {
			case <-finished:
			case <-time.After(3 * time.Second):
				t.Fatal("handler did not exit")
			}
			if err := <-result; (err != nil) != timeout {
				t.Fatalf("client error = %v, timeout = %v", err, timeout)
			}
		})
	}
}

func TestGRPCTLSListenerAndShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	service, err := New(ctx, ExternalServices{})
	require.NoError(t, err)
	cert, err := tls.LoadX509KeyPair("external/cert/localhost+1.pem", "external/cert/localhost+1-key.pem")
	require.NoError(t, err)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	listener = tls.NewListener(listener, &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12, NextProtos: []string{"h2"}})
	done := make(chan error, 1)
	go func() { done <- service.serveGRPC(ctx, listener) }()
	defer service.GRPCServer().Stop()
	pem, err := os.ReadFile("external/cert/localhost+1.pem")
	require.NoError(t, err)
	roots := x509.NewCertPool()
	require.True(t, roots.AppendCertsFromPEM(pem))
	conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{RootCAs: roots, ServerName: "localhost"})))
	require.NoError(t, err)
	defer conn.Close()
	callCtx, callCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer callCancel()
	response, err := grpcAPI.NewQRCodeServiceClient(conn).GenerateQRCode(callCtx, &grpcAPI.GenerateQRCodeRequest{Message: &grpcAPI.MessageRequest{Text: "hello"}})
	require.NoError(t, err)
	require.Equal(t, "image/png", response.ContentType)
	require.NotEmpty(t, response.Image)
	cancel()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("gRPC listener did not shut down")
	}
	// The listening socket must have been released.
	reopened, err := net.Listen("tcp", listener.Addr().String())
	require.NoError(t, err)
	reopened.Close()
}

func TestServiceRunReportsGRPCListenerFailure(t *testing.T) {
	cert, err := os.ReadFile("external/cert/localhost+1.pem")
	require.NoError(t, err)
	key, err := os.ReadFile("external/cert/localhost+1-key.pem")
	require.NoError(t, err)
	t.Setenv("TLS_CERT", string(cert))
	t.Setenv("TLS_KEY", string(key))
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer listener.Close()
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	service, err := New(ctx, ExternalServices{})
	require.NoError(t, err)
	err = service.Run(ctx, "0", "0", "0", "0", port)
	require.ErrorContains(t, err, "listening for gRPC")
	require.NoError(t, ctx.Err())
}
