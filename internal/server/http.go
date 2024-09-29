package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"server/config"
	"server/internal/handlers"
	"server/internal/middleware"
	"server/internal/repository"
	"server/internal/service"

	dice "github.com/dicedb/go-dice"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
}

func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Run() error {
	// Step1: Initialize DiceDB client
	diceClient, err := initDiceClient(s.config)
	if err != nil {
		return fmt.Errorf("failed to initialize DiceDB client: %w", err)
	}

	// Step2: Initialize repository
	repo := repository.NewRepository(diceClient)

	// Step3: Initialize service
	svc := service.NewService(repo)

	// Step4: Initialize handler
	handler := handlers.NewHandler(svc)

	// TODO: We should ideally move this to routes package
	// Step5: Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/cli/{cmd}", handler.CLIHandler)
	mux.HandleFunc("/search", handler.Search)

	handlerMux := &HandlerMux{
		mux: mux,
		rateLimiter: func(w http.ResponseWriter, r *http.Request, next http.Handler) {
			middleware.RateLimiter(diceClient, next, s.config.RequestLimit, s.config.RequestWindow).ServeHTTP(w, r)
		},
	}

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    s.config.ServerPort,
		Handler: handlerMux,
	}

	// Start server
	go func() {
		log.Printf("Starting server on %s", s.config.ServerPort)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait until the timeout deadline
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited properly")
	return nil
}

// HandlerMux wraps ServeMux and forces REST paths to lowercase
// and attaches a rate limiter with the handler
type HandlerMux struct {
	mux         *http.ServeMux
	rateLimiter func(http.ResponseWriter, *http.Request, http.Handler)
}

func (hm *HandlerMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Convert the path to lowercase before passing to the underlying mux.
	r.URL.Path = strings.ToLower(r.URL.Path)
	hm.rateLimiter(w, r, hm.mux)
}

// initDiceClient initializes dice client
func initDiceClient(config *config.Config) (*dice.Client, error) {
	diceClient := dice.NewClient(&dice.Options{
		Addr:        config.DiceAddr,
		DialTimeout: 10 * time.Second,
		MaxRetries:  10,
	})

	// Ping the dice client to verify the connection
	err := diceClient.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}

	return diceClient, nil
}
