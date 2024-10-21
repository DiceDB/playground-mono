package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"server/config"
	"server/internal/db"
	"server/internal/server"
	"sync"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	configValue := config.LoadConfig()
	diceDBAdminClient, err := db.InitDiceClient(configValue, true)
	if err != nil {
		slog.Error("Failed to initialize DiceDB Admin client: %v", slog.Any("err", err))
		os.Exit(1)
	}

	diceDBClient, err := db.InitDiceClient(configValue, false)
	if err != nil {
		slog.Error("Failed to initialize DiceDB client: %v", slog.Any("err", err))
		os.Exit(1)
	}

	// Graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	// Register a cleanup manager, this runs user DiceDB instance cleanup job at configured frequency
	cleanupManager := server.NewCleanupManager(diceDBAdminClient, diceDBClient, configValue.Server.CronCleanupFrequency)
	go func() {
		defer wg.Done()
		cleanupManager.Run(ctx)
	}()

	// Create mux and register routes
	mux := http.NewServeMux()
	httpServer := server.NewHTTPServer(":8080", mux, diceDBAdminClient, diceDBClient, configValue.Server.RequestLimitPerMin,
		configValue.Server.RequestWindowSec)
	mux.HandleFunc("/health", httpServer.HealthCheck)
	mux.HandleFunc("/shell/exec/{cmd}", httpServer.CliHandler)
	mux.HandleFunc("/search", httpServer.SearchHandler)

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Run the HTTP Server
		if err := httpServer.Run(ctx); err != nil {
			slog.Error("server failed: %v\n", slog.Any("err", err))
			diceDBAdminClient.CloseDiceDB()
			diceDBClient.CloseDiceDB()
			cancel()
		}
	}()

	wg.Wait()
	slog.Info("Server has shut down gracefully")
}
