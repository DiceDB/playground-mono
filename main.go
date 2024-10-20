package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"server/config"
	"server/internal/cron"
	"server/internal/db"
	"server/internal/server"
)

func main() {
	configValue := config.LoadConfig()

	// sys dice client for doing server specific operations
	sysDiceClient, err := db.InitDiceClient(configValue.SysDiceDBAddr)
	if err != nil {
		slog.Error("Failed to initialize DiceDB client: %v", slog.Any("err", err))
		os.Exit(1)
	}

	//  user demo dice client for handling user demo operations
	userDemoDiceClient, err := db.InitDiceClient(configValue.UserDemoDiceDBAddr)
	if err != nil {
		slog.Error("Failed to initialize DiceDB client: %v", slog.Any("err", err))
		os.Exit(1)
	}

	// Create mux and register routes
	mux := http.NewServeMux()
	httpServer := server.NewHTTPServer(":8080", mux, userDemoDiceClient, sysDiceClient, configValue.RequestLimitPerMin, configValue.RequestWindowSec)
	mux.HandleFunc("/health", httpServer.HealthCheck)
	mux.HandleFunc("/shell/exec/{cmd}", httpServer.CliHandler)
	mux.HandleFunc("/search", httpServer.SearchHandler)

	// Graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cron.StartCleanupCron(ctx, userDemoDiceClient, sysDiceClient, configValue.CleanupCronFrequency)

	// Run the HTTP Server
	if err := httpServer.Run(ctx); err != nil {
		slog.Error("server failed: %v\n", slog.Any("err", err))
	}

}
