package main

import (
	"log/slog"
	"os"
)

func InitLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Strip the date, keep just the time for visual clarity
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				return slog.String(slog.TimeKey, t.Format("15:04:05"))
			}
			return a
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	slog.Info("User logged in", "user_id", 42)
}
