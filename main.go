package main

import (
	"context"
	"log/slog"
	"net/http"
	"server/config"
	"server/internal/db"
	"server/internal/server"
)

func main() {
	configValue := config.LoadConfig()
	diceClient, err := db.InitDiceClient(configValue)
	if err != nil {
		slog.Error("Failed to initialize DiceDB client: %v", slog.Any("err", err))
	}

	// Create mux and register routes
	mux := http.NewServeMux()
	httpServer := server.NewHTTPServer(":8080", mux, diceClient, configValue.RequestLimitPerMin, configValue.RequestWindowSec)
	mux.HandleFunc("/health", httpServer.HealthCheck)
	mux.HandleFunc("/shell/exec/{cmd}", httpServer.CliHandler)
	mux.HandleFunc("/search", httpServer.SearchHandler)

	// Graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run the HTTP Server
	if err := httpServer.Run(ctx); err != nil {
		slog.Error("server failed: %v\n", slog.Any("err", err))
	}
}
