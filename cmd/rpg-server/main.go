package main

import (
	"github.com/bornostojak/repugnant/internal/store"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bornostojak/repugnant/internal/httpapi"
)

func main() {
	level := new(slog.LevelVar)
	level.Set(logLevel(os.Getenv("RPG_LOG_LEVEL")))
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	address := os.Getenv("RPG_HTTP_ADDR")
	if address == "" {
		address = "0.0.0.0:8080"
	}
	driver, dsn := os.Getenv("RPG_DB_DRIVER"), os.Getenv("RPG_DB_DSN")
	database, err := store.Open(driver, dsn)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()
	logger.Info("starting rpg server", "address", address)
	handler := httpapi.NewWithStore(logger, database).Handler()
	if webDir := os.Getenv("RPG_WEB_DIR"); webDir != "" {
		handler = httpapi.WithWeb(handler, filepath.Clean(webDir))
	}
	if err := http.ListenAndServe(address, handler); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

// logLevel honors RPG_LOG_LEVEL (error, warn, info, debug; case-insensitive)
// as TASTE.md requires level-selectable logging. Info is the safe default
// when the variable is unset or unrecognized.
func logLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
