package sse_test

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"

	commonHTTP "htqrcode/common/http"
	"htqrcode/common/sse"
)

func TestStreamFlushesAndStops(t *testing.T) {
	for _, closeChannel := range []bool{false, true} {
		t.Run(map[bool]string{false: "disconnect", true: "channel closed"}[closeChannel], func(t *testing.T) {
			e := commonHTTP.NewEcho()
			events := make(chan sse.Event, 1)
			done := make(chan error, 1)
			e.GET("/sse/events", func(c *echo.Context) error {
				err := sse.Stream(c, events)
				done <- err
				return err
			})
			server := httptest.NewServer(e)
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/sse/events", nil)
			response, err := server.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			for key, want := range map[string]string{"Content-Type": "text/event-stream", "Cache-Control": "no-cache", "X-Accel-Buffering": "no"} {
				if got := response.Header.Get(key); got != want {
					t.Fatalf("%s = %q, want %q", key, got, want)
				}
			}
			reader := bufio.NewReader(response.Body)
			readFrame(t, reader, ": connected\n\n")
			events <- sse.Event{ID: []byte("42"), Event: []byte("updated"), Data: []byte("hello")}
			readFrame(t, reader, "id: 42\ndata: hello\nevent: updated\n\n")
			if closeChannel {
				close(events)
			} else {
				cancel()
			}
			select {
			case err := <-done:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("stream did not stop")
			}
		})
	}
}

func TestStreamHeartbeatSurvivesOrdinaryRequestTimeout(t *testing.T) {
	e := commonHTTP.NewEcho()
	done := make(chan error, 1)
	e.GET("/sse/events", func(c *echo.Context) error {
		err := sse.Stream(c, nil)
		done <- err
		return err
	})
	server := httptest.NewServer(e)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/sse/events", nil)
	response, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	reader := bufio.NewReader(response.Body)
	readFrame(t, reader, ": connected\n\n")
	// The heartbeat at 15 seconds proves that the 10-second API timeout and
	// per-write deadline do not limit the lifetime of an idle SSE connection.
	readFrame(t, reader, ": heartbeat\n\n")
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("heartbeat stream did not stop")
	}
}

func readFrame(t *testing.T, reader *bufio.Reader, want string) {
	t.Helper()
	frame := make([]byte, len(want))
	if _, err := io.ReadFull(reader, frame); err != nil {
		t.Fatal(err)
	}
	if string(frame) != want {
		t.Fatalf("frame = %q, want %q", frame, want)
	}
}

type flushErrorWriter struct {
	header http.Header
	err    error
}

func (w *flushErrorWriter) Header() http.Header          { return w.header }
func (*flushErrorWriter) WriteHeader(int)                {}
func (*flushErrorWriter) Write(data []byte) (int, error) { return len(data), nil }
func (w *flushErrorWriter) FlushError() error            { return w.err }

func TestStreamPropagatesFlushError(t *testing.T) {
	e := commonHTTP.NewEcho()
	want := errors.New("flush failed")
	var got error
	e.GET("/sse/events", func(c *echo.Context) error {
		got = sse.Stream(c, nil)
		return got
	})
	e.ServeHTTP(&flushErrorWriter{header: make(http.Header), err: want}, httptest.NewRequest(http.MethodGet, "/sse/events", nil))
	if !errors.Is(got, want) {
		t.Fatalf("error = %v, want %v", got, want)
	}
}
