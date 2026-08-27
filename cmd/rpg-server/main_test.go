package main

import (
	"log/slog"
	"testing"
)

func TestLogLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"":        slog.LevelInfo,
		"info":    slog.LevelInfo,
		"bogus":   slog.LevelInfo,
		"debug":   slog.LevelDebug,
		"DEBUG":   slog.LevelDebug,
		"warn":    slog.LevelWarn,
		"warning": slog.LevelWarn,
		" error ": slog.LevelError,
	}
	for input, want := range cases {
		if got := logLevel(input); got != want {
			t.Errorf("logLevel(%q) = %v, want %v", input, got, want)
		}
	}
}
