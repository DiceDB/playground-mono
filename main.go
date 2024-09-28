package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"server/config"
	"sync"
	"time"

	"server/internal/api"
	"server/internal/db"
	"server/internal/middleware"

	dice "github.com/dicedb/go-dice"
)

type HTTPServer struct {
	httpServer *http.Server
	diceClient *dice.Client
}

func NewHTTPServer(addr string, mux *http.ServeMux, client *dice.Client) *HTTPServer {
	return &HTTPServer{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		diceClient: client,
	}
}

func initDiceClient(configValue *config.Config) (*dice.Client, error) {
	client := dice.NewClient(&dice.Options{
		Addr:        configValue.DiceAddr,
		DialTimeout: 10 * time.Second,
		MaxRetries:  10,
	})

	// Ping the dice client to verify the connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

func (s *HTTPServer) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Starting server at %s\n", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")
	return s.Shutdown()
}

func (s *HTTPServer) Shutdown() error {
	// Additional cleanup if necessary
	if err := s.diceClient.Close(); err != nil {
		log.Printf("Failed to close dice client: %v", err)
	}
	return s.httpServer.Shutdown(context.Background())
}

func main() {
	configValue := config.LoadConfig()
	diceClient, err := initDiceClient(configValue)
	if err != nil {
		log.Fatalf("Failed to initialize dice client: %v", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/", middleware.RateLimiter(diceClient, http.HandlerFunc(api.HealthCheck), configValue.RequestLimit, configValue.RequestWindow))
	api.RegisterRoutes(mux)

	httpServer := NewHTTPServer(":8080", mux, diceClient)

	// context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// run the Http Server
	if err := httpServer.Run(ctx); err != nil {
		log.Printf("Server failed: %v\n", err)
	}
}
