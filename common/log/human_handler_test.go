package log

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestHumanHandlerFormatsRecordsAndAttributes(t *testing.T) {
	var output bytes.Buffer
	handler := NewHandler(&output, &Options{
		HandlerOptions: &slog.HandlerOptions{Level: slog.LevelDebug},
		TimeFormat:     "15:04:05",
		NoColor:        true,
		SortKeys:       true,
	})

	record := slog.NewRecord(time.Date(2026, 7, 29, 10, 11, 12, 0, time.UTC), slog.LevelInfo, "request complete", 0)
	record.AddAttrs(slog.String("route", "/health"), slog.Int("status", 200))
	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	got := output.String()
	for _, want := range []string{"10:11:12", " INFO ", "request complete", "route=/health", "status=200"} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q does not contain %q", got, want)
		}
	}
}

func TestHumanHandlerPreservesGroupsAndLevelFiltering(t *testing.T) {
	var output bytes.Buffer
	handler := NewHandler(&output, &Options{
		HandlerOptions: &slog.HandlerOptions{Level: slog.LevelWarn},
		NoColor:        true,
	})

	if handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("info level should be disabled")
	}

	logger := slog.New(handler.WithGroup("request").WithAttrs([]slog.Attr{slog.String("id", "abc")}))
	logger.Warn("failed", slog.String("reason", "timeout"))

	got := output.String()
	for _, want := range []string{" WARN ", "request.id=abc", "request.reason=timeout"} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q does not contain %q", got, want)
		}
	}
}
