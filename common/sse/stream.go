package sse

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

const (
	heartbeatInterval = 15 * time.Second
	writeTimeout      = 10 * time.Second
)

// Stream sends events until the request is canceled or events is closed. A nil
// channel sends only heartbeats. The caller owns subscription cleanup and must
// authenticate and validate the request before calling Stream.
// It provides live delivery only; replay and durable fan-out belong to the module.
func Stream(c *echo.Context, events <-chan Event) error {
	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	// Echo's Flush method discards underlying flush errors. Use its underlying
	// writer for the controller while keeping writes on Echo for status tracking.
	response, err := echo.UnwrapResponse(w)
	if err != nil {
		return err
	}
	controller := http.NewResponseController(response.ResponseWriter)
	send := func(event Event) error {
		// Bound each write to slow clients while leaving the stream lifetime open.
		if err := controller.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil && !errors.Is(err, http.ErrNotSupported) {
			return err
		}
		if err := event.MarshalTo(w); err != nil {
			return err
		}
		return controller.Flush()
	}
	defer controller.SetWriteDeadline(time.Time{})
	if err := send(Event{Comment: []byte("connected")}); err != nil {
		return err
	}
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.Request().Context().Done():
			return nil
		case event, ok := <-events:
			if !ok {
				return nil
			}
			if err := send(event); err != nil {
				return err
			}
		case <-ticker.C:
			if err := send(Event{Comment: []byte("heartbeat")}); err != nil {
				return err
			}
		}
	}
}
