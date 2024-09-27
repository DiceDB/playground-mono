package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"server/internal/api"
	"server/internal/middleware"
)

type HTTPServer struct {
	httpServer *http.Server
}

func NewHTTPServer(addr string, mux *http.ServeMux) *HTTPServer {
	return &HTTPServer{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
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
	return s.httpServer.Shutdown(context.Background())
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", middleware.RateLimiter(http.HandlerFunc(api.HealthCheck)))
	api.RegisterRoutes(mux)

	httpServer := NewHTTPServer(":8080", mux)

	// context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// run the Http Server
	if err := httpServer.Run(ctx); err != nil {
		log.Printf("Server failed: %v\n", err)
	}
}
