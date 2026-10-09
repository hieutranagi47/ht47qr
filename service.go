package htqrcode

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	echo "github.com/labstack/echo/v5"
	"golang.org/x/sync/errgroup"

	"htqrcode/client"
	grpcTransport "htqrcode/common/grpc"
	commonHTTP "htqrcode/common/http"
	"htqrcode/common/log"
	"htqrcode/common/module"
	"htqrcode/common/module/contracts"
	"htqrcode/qrcode"
	"htqrcode/shorten_url"
)

// FileStorage stores files and returns their public URL.
type FileStorage interface {
	StoreFile(ctx context.Context, path string, content []byte) (string, error)
}

// ExternalServices provides external API service dependencies.
// For production, use real HTTP clients. For tests, inject stubs.
type ExternalServices struct {
	FileStorage FileStorage
	// Database enables SQLite-backed modules. The caller owns its lifetime.
	Database *sql.DB
}

type Service struct {
	echoRouter    *echo.Echo
	echoSSERouter *echo.Echo
	grpcServer    *grpcTransport.Server

	modules []module.Module
}

func New(
	ctx context.Context,
	services ExternalServices,
) (Service, error) {
	e := commonHTTP.NewEcho()
	sse := commonHTTP.NewEcho()

	// We use a pointer here so modules can register their contracts during Init(),
	// then all modules can call each other after initialization completes.
	moduleContracts := &contracts.Contracts{}

	modules := []module.Module{qrcode.NewModule()}
	if services.Database != nil {
		modules = append(modules, shorten_url.NewModule(services.Database))
	}
	modules = append(modules, client.NewModule())

	for _, module := range modules {
		start := time.Now()

		if err := module.Init(ctx); err != nil {
			return Service{}, fmt.Errorf("initializing module %s failed: %w", module.Name(), err)
		}

		if err := module.RegisterContracts(ctx, moduleContracts); err != nil {
			return Service{}, fmt.Errorf("registering module %s failed: %w", module.Name(), err)
		}

		log.FromContext(ctx).With(
			"duration", time.Since(start),
			"module", module.Name(),
		).Debug("Initialized module")
	}

	if err := moduleContracts.Verify(); err != nil {
		return Service{}, fmt.Errorf("verifying module contracts failed: %w", err)
	}

	grpcServer := grpcTransport.NewServer()

	for _, m := range modules {
		err := m.RegisterHttp(ctx, e)
		if err != nil {
			grpcServer.Stop()
			return Service{}, fmt.Errorf("registering http for module %s failed: %w", m.Name(), err)
		}
		if provider, ok := m.(module.GRPCModule); ok {
			if err := provider.RegisterGRPC(ctx, grpcServer); err != nil {
				grpcServer.Stop()
				return Service{}, fmt.Errorf("registering gRPC for module %s failed: %w", m.Name(), err)
			}
		}
		if err := m.RegisterSSE(ctx, sse.Group("/sse")); err != nil {
			grpcServer.Stop()
			return Service{}, fmt.Errorf("registering SSE for module %s failed: %w", m.Name(), err)
		}
	}

	return Service{
		echoRouter:    e,
		grpcServer:    grpcServer,
		echoSSERouter: sse,
		modules:       modules,
	}, nil
}

func (s *Service) Run(ctx context.Context, httpPort, httpsPort, ssePort, ssesPort string, grpcPorts ...string) (runErr error) {
	if len(grpcPorts) > 1 {
		return fmt.Errorf("at most one gRPC port may be configured")
	}
	cert := []byte(os.Getenv("TLS_CERT"))
	key := []byte(os.Getenv("TLS_KEY"))
	if len(cert) == 0 || len(key) == 0 {
		return fmt.Errorf("TLS_CERT and TLS_KEY must contain PEM-encoded certificate and private key")
	}
	if _, err := tls.X509KeyPair(cert, key); err != nil {
		return fmt.Errorf("loading TLS_CERT/TLS_KEY (expected PEM content with real newlines, not file paths or literal \\n): %w", err)
	}
	for _, port := range []string{httpPort, httpsPort, ssePort, ssesPort} {
		if port == "" {
			return fmt.Errorf("HTTP, HTTPS, SSE and SSE TLS ports must be configured")
		}
	}
	g, ctx := errgroup.WithContext(ctx)

	// Separate listeners keep the streaming surface isolated from ordinary APIs.
	for _, listener := range []struct {
		port   string
		tls    bool
		stream bool
	}{
		{port: httpPort},
		{port: httpsPort, tls: true},
		{port: ssePort, stream: true},
		{port: ssesPort, tls: true, stream: true},
	} {
		router := s.echoRouter
		if listener.stream {
			router = s.echoSSERouter
		}
		g.Go(func() error {
			if ctx.Err() != nil {
				return nil
			}
			config := serverConfig(ctx, listener.port, listener.stream)
			return serveHTTP(ctx, config, router, listener.tls, cert, key)
		})
	}

	if len(grpcPorts) == 1 && grpcPorts[0] != "" {
		g.Go(func() error { return s.runGRPC(ctx, grpcPorts[0], cert, key) })
	}
	return g.Wait()
}

func serveHTTP(ctx context.Context, config echo.StartConfig, router *echo.Echo, secure bool, cert, key []byte) error {
	var server *http.Server
	var shutdownErr error
	beforeServe := config.BeforeServeFunc
	config.BeforeServeFunc = func(s *http.Server) error {
		server = s
		if beforeServe != nil {
			return beforeServe(s)
		}
		return nil
	}
	onShutdownError := config.OnShutdownError
	config.OnShutdownError = func(err error) {
		// Shutdown only stops accepting connections when its deadline expires.
		// Close the remaining connections and cancel their request contexts.
		shutdownErr = errors.Join(err, server.Close())
		if onShutdownError != nil {
			onShutdownError(shutdownErr)
		}
	}
	var err error
	if secure {
		// Echo treats byte slices as PEM content rather than file paths.
		err = config.StartTLS(ctx, router, cert, key)
	} else {
		err = config.Start(ctx, router)
	}
	// Echo waits for its shutdown goroutine before Start/StartTLS returns.
	return errors.Join(err, shutdownErr)
}

func serverConfig(ctx context.Context, port string, stream bool) echo.StartConfig {
	return echo.StartConfig{
		Address:         ":" + port,
		GracefulTimeout: 15 * time.Second,
		HideBanner:      true,
		BeforeServeFunc: func(server *http.Server) error {
			server.ReadHeaderTimeout = 5 * time.Second
			if stream {
				server.WriteTimeout = 0
				// Streaming handlers must exit as soon as shutdown starts.
				server.BaseContext = func(net.Listener) context.Context { return ctx }
			} else {
				// Preserve context values while allowing ordinary requests to drain.
				server.BaseContext = func(net.Listener) context.Context { return context.WithoutCancel(ctx) }
			}
			return nil
		},
		OnShutdownError: func(err error) {
			log.FromContext(ctx).Error("shutting down HTTP server failed", "error", err, "port", port)
		},
	}
}

// HTTPHandler exposes the ordinary API router for embedding and component tests.
func (s *Service) HTTPHandler() http.Handler { return s.echoRouter }

// SSEHandler exposes the streaming router, which is served on separate listeners.
func (s *Service) SSEHandler() http.Handler { return s.echoSSERouter }

// GRPCServer exposes the registered server for embedding and component tests.
// Callers that serve it directly own listener security and shutdown.
func (s *Service) GRPCServer() *grpcTransport.Server { return s.grpcServer }

func (s *Service) runGRPC(ctx context.Context, port string, certPEM, keyPEM []byte) error {
	if ctx.Err() != nil {
		return nil
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return fmt.Errorf("loading gRPC TLS certificate: %w", err)
	}
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("listening for gRPC: %w", err)
	}
	return s.serveGRPC(ctx, tls.NewListener(listener, &tls.Config{
		Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12, NextProtos: []string{"h2"},
	}))
}

func (s *Service) serveGRPC(ctx context.Context, listener net.Listener) error {
	defer listener.Close()
	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		select {
		case <-ctx.Done():
		case <-done:
		}
		s.grpcServer.Shutdown(15 * time.Second)
	}()
	err := s.grpcServer.Serve(listener)
	close(done)
	<-stopped
	return err
}
