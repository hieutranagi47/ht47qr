package http

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/lithammer/shortuuid/v3"

	"htqrcode/common/log"
)

const (
	TestNameHeader          = "TestName"
	CorrelationIDHttpHeader = "Correlation-ID"
)

func useMiddlewares(e *echo.Echo) {
	e.Use(
		middleware.CORSWithConfig(middleware.CORSConfig{
			UnsafeAllowOriginFunc: allowSameOrigin,
			AllowMethods: []string{
				http.MethodGet,
				http.MethodHead,
				http.MethodPost,
			},
		}),
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20.0)),
		middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
			Timeout: 10 * time.Second,
			Skipper: isSSERequest,
		}),
		middleware.Recover(),
		// Correlation-ID runs first: available in context for the request log middleware.
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				req := c.Request()
				ctx := req.Context()

				reqCorrelationID := req.Header.Get(CorrelationIDHttpHeader)
				if reqCorrelationID == "" {
					reqCorrelationID = shortuuid.New()
				}

				logger := slog.With("correlation_id", reqCorrelationID)

				if testName := c.Request().Header.Get("TestName"); testName != "" {
					logger = logger.With("test_name", testName)
				}

				ctx = log.ToContext(ctx, logger)
				ctx = log.ContextWithCorrelationID(ctx, reqCorrelationID)
				c.SetRequest(req.WithContext(ctx))
				c.Response().Header().Set(CorrelationIDHttpHeader, reqCorrelationID)

				return next(c)
			}
		},
		requestLogMiddleware,
		sseMiddleware,
	)
}

func allowSameOrigin(c *echo.Context, origin string) (string, bool, error) {
	scheme := c.Scheme()
	expectedOrigin := scheme + "://" + c.Request().Host
	// Browsers omit the default port when serializing an Origin.
	defaultPort := ":80"
	if scheme == "https" {
		defaultPort = ":443"
	}
	if !strings.EqualFold(strings.TrimSuffix(origin, defaultPort), strings.TrimSuffix(expectedOrigin, defaultPort)) {
		return "", false, echo.NewHTTPError(http.StatusForbidden, "cross-origin requests are not allowed")
	}
	return origin, true, nil
}

type bodyCapturingWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *bodyCapturingWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyCapturingWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyCapturingWriter) Flush() {
	err := http.NewResponseController(w.ResponseWriter).Flush()
	if err != nil && !errors.Is(err, http.ErrNotSupported) {
		slog.Warn("response writer flush failed", "error", err)
	}
}

func (w *bodyCapturingWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(w.ResponseWriter).Hijack()
}

func (w *bodyCapturingWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func requestLogMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// Never buffer an unbounded stream for request logging.
		if isSSERequest(c) {
			return next(c)
		}
		// Read request body and restore it for the handler.
		var reqBody []byte
		if c.Request().Body != nil && !strings.HasPrefix(c.Request().Header.Get("Content-Type"), "multipart/form-data") {
			reqBody, _ = io.ReadAll(c.Request().Body)
		}
		if reqBody != nil {
			c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// Capture response body via MultiWriter.
		resBody := new(bytes.Buffer)
		response := c.Response()
		c.SetResponse(&bodyCapturingWriter{
			Writer:         io.MultiWriter(response, resBody),
			ResponseWriter: response,
		})

		start := time.Now()
		err := next(c)
		duration := time.Since(start)

		ctx := c.Request().Context()

		logger := log.FromContext(ctx).With(
			"URI", c.Request().RequestURI,
			"status", responseStatus(c.Response(), err),
			"method", c.Request().Method,
			"duration", duration.String(),
		)
		if err != nil {
			logger = logger.With("error", err)
		}
		logger = logger.With("request_body", truncateBodyForLog(string(reqBody)))

		body := resBody.String()
		if utf8.ValidString(body) {
			if isDebug := log.FromContext(ctx).Enabled(ctx, slog.LevelDebug); !isDebug {
				body = truncateBodyForLog(body)
			}
			logger = logger.With("response_body", body)
		} else {
			logger = logger.With("response_body", "<binary data>")
		}

		logger.Info("Request done")
		return err
	}
}

func responseStatus(response http.ResponseWriter, err error) int {
	_, status := echo.ResolveResponseStatus(response, err)
	return status
}

func isSSERequest(c *echo.Context) bool {
	return strings.HasPrefix(c.Path(), "/sse/")
}

func sseMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if !isSSERequest(c) {
			return next(c)
		}
		logger := log.FromContext(c.Request().Context())
		logger.Info("SSE client connected", "ip", c.RealIP())
		defer logger.Info("SSE client disconnected", "ip", c.RealIP())
		// Stream handlers set headers after authentication and validation succeed.
		return next(c)
	}
}
