package sse

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestEventFraming(t *testing.T) {
	for _, test := range []struct {
		name  string
		event Event
		want  string
	}{
		{"empty", Event{}, ""},
		{"message without cursor reset", Event{Data: []byte("hello")}, "data: hello\n\n"},
		{"empty message", Event{Data: []byte{}}, "data: \n\n"},
		{"multiline", Event{ID: []byte("42"), Data: []byte("one\r\ntwo\rthree\n"), Event: []byte("updated"), Retry: []byte("1000")}, "id: 42\ndata: one\ndata: two\ndata: three\ndata: \nevent: updated\nretry: 1000\n\n"},
		{"comment", Event{Comment: []byte("alive\r\nagain")}, ": alive\n: again\n\n"},
		{"retry only", Event{Retry: []byte("2000")}, "retry: 2000\n\n"},
		{"reset cursor", Event{ID: []byte{}}, "id: \n\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := test.event.MarshalTo(&output); err != nil {
				t.Fatal(err)
			}
			if output.String() != test.want {
				t.Fatalf("frame = %q, want %q", output.String(), test.want)
			}
		})
	}
}

func TestEventRejectsFieldInjectionBeforeWriting(t *testing.T) {
	for _, event := range []Event{
		{ID: []byte("one\nevent: injected"), Data: []byte("hello")},
		{ID: []byte("one\x00")},
		{Event: []byte("one\rdata: injected")},
		{Retry: []byte("-1")},
		{Retry: []byte("10\n\ndata: injected")},
	} {
		var output bytes.Buffer
		if err := event.MarshalTo(&output); err == nil || output.Len() != 0 {
			t.Fatalf("invalid event wrote %q, error = %v", output.String(), err)
		}
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestEventPropagatesWriteFailure(t *testing.T) {
	event := Event{Data: []byte("hello")}
	if err := event.MarshalTo(failingWriter{io.ErrClosedPipe}); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("error = %v, want closed pipe", err)
	}
}
