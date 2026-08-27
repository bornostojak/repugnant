package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/bornostojak/repugnant/internal/httpapi"
)

func main() {
	level := new(slog.LevelVar)
	level.Set(slog.LevelInfo)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	address := os.Getenv("RPG_HTTP_ADDR")
	if address == "" {
		address = ":8080"
	}
	logger.Info("starting rpg server", "address", address)
	if err := http.ListenAndServe(address, httpapi.New(logger).Handler()); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
