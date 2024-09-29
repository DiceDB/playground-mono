package main

import (
	"context"
	"log"
	"net/http"
	"server/config"
	"server/internal/db"
	"server/internal/server" // Import the new package for HTTPServer
)

func main() {
	configValue := config.LoadConfig()
	diceClient, err := db.InitDiceClient(configValue)
	if err != nil {
		log.Fatalf("Failed to initialize dice client: %v", err)
	}

	// Create mux and register routes
	mux := http.NewServeMux()
	httpServer := server.NewHTTPServer(":8080", mux, diceClient, configValue.RequestLimit, configValue.RequestWindow)
	mux.HandleFunc("/health", httpServer.HealthCheck)
	mux.HandleFunc("/cli/{cmd}", httpServer.CliHandler)
	mux.HandleFunc("/search", httpServer.SearchHandler)

	// Graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run the HTTP Server
	if err := httpServer.Run(ctx); err != nil {
		log.Printf("Server failed: %v\n", err)
	}
}
