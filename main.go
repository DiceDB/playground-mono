package main

import (
	"log"
	"server/config"
	"server/internal/server"
)

func main() {
	configValue := config.Load()
	srv := server.NewServer(configValue)

	if err := srv.Run(); err != nil {
		// Crash now
		log.Fatalf("Server error: %v", err)
	}
}
