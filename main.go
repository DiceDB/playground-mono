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
	"server/internal/middleware"

	redis "github.com/dicedb/go-dice"
)

type HTTPServer struct {
	httpServer *http.Server
	diceClient *redis.Client
}

func NewHTTPServer(addr string, mux *http.ServeMux, client *redis.Client) *HTTPServer {
	return &HTTPServer{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		diceClient: client,
	}
}

func initDiceClient(config *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:        config.RedisAddr,
		DialTimeout: 10 * time.Second,
		MaxRetries:  10,
	})

	// Ping the Redis server to verify the connection
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
		log.Printf("Failed to close Redis client: %v", err)
	}
	return s.httpServer.Shutdown(context.Background())
}

func main() {
	config := config.LoadConfig()
	diceClient, err := initDiceClient(config)
	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/", middleware.RateLimiter(diceClient, http.HandlerFunc(api.HealthCheck), config.RequestLimit, config.RequestWindow))
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
