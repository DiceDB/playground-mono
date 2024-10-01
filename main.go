package main

import (
	"log"

	"github.com/DiceDB/playground-mono/config"
	"github.com/DiceDB/playground-mono/internal/server"
)

func main() {
	configValue := config.Load()
	srv := server.NewServer(configValue)

	if err := srv.Run(); err != nil {
		// Crash now
		log.Fatalf("Server error: %v", err)
	}
}
