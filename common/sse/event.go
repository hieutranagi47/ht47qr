package sse

import (
	"bytes"
	"fmt"
	"io"
)

// Event represents a Server-Sent Event. Nil fields are omitted; an empty,
// non-nil Data emits an empty message and an empty ID resets the replay cursor.
// Retry contains an ASCII integer in milliseconds.
type Event struct {
	ID      []byte
	Data    []byte
	Event   []byte
	Retry   []byte
	Comment []byte
}

// MarshalTo writes one complete event. Validate before writing so invalid field
// values cannot inject additional SSE frames or leave a partial event behind.
func (ev *Event) MarshalTo(w io.Writer) error {
	if bytes.ContainsAny(ev.ID, "\r\n\x00") || bytes.ContainsAny(ev.Event, "\r\n") {
		return fmt.Errorf("SSE ID and event name must be single-line values; ID must not contain NUL")
	}
	for _, digit := range ev.Retry {
		if digit < '0' || digit > '9' {
			return fmt.Errorf("SSE retry must be an integer in milliseconds")
		}
	}
	var frame bytes.Buffer
	if ev.ID != nil {
		fmt.Fprintf(&frame, "id: %s\n", ev.ID)
	}
	if ev.Data != nil {
		writeLines(&frame, "data: ", ev.Data)
	}
	if ev.Event != nil {
		fmt.Fprintf(&frame, "event: %s\n", ev.Event)
	}
	if len(ev.Retry) > 0 {
		fmt.Fprintf(&frame, "retry: %s\n", ev.Retry)
	}
	if ev.Comment != nil {
		writeLines(&frame, ": ", ev.Comment)
	}
	if frame.Len() == 0 {
		return nil
	}
	frame.WriteByte('\n')
	_, err := frame.WriteTo(w)
	return err
}

func writeLines(w io.Writer, prefix string, value []byte) {
	// SSE recognizes CRLF, CR, and LF as line endings.
	value = bytes.ReplaceAll(value, []byte("\r\n"), []byte("\n"))
	value = bytes.ReplaceAll(value, []byte("\r"), []byte("\n"))
	for _, line := range bytes.Split(value, []byte("\n")) {
		fmt.Fprintf(w, "%s%s\n", prefix, line)
	}
}
