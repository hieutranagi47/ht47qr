package log

import (
	"log/slog"
	"os"
)

func Init(level slog.Level) {
  opts := &Options{
    HandlerOptions: &slog.HandlerOptions{
      Level: level,
    },
    TimeFormat: "[15:04:05.000]",
  }

  logger := slog.New(NewHandler(os.Stderr, opts))
  slog.SetDefault(logger)
}
