package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"server/internal/constants"
	"server/internal/db"
	"server/internal/middleware"
	util "server/pkg/util"
	"strings"
	"sync"
	"time"
)

type HTTPServer struct {
	httpServer *http.Server
	DiceClient *db.DiceDB
}

// HandlerMux wraps ServeMux and forces REST paths to lowercase
// and attaches a rate limiter with the handler
type HandlerMux struct {
	mux         *http.ServeMux
	rateLimiter func(http.ResponseWriter, *http.Request, http.Handler)
}

func (cim *HandlerMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Convert the path to lowercase before passing to the underlying mux.
	r.URL.Path = strings.ToLower(r.URL.Path)
	// Apply rate limiter
	cim.rateLimiter(w, r, cim.mux)
}

func NewHTTPServer(addr string, mux *http.ServeMux, client *db.DiceDB, limit, window int) *HTTPServer {
	handlerMux := &HandlerMux{
		mux: mux,
		rateLimiter: func(w http.ResponseWriter, r *http.Request, next http.Handler) {
			middleware.RateLimiter(client, next, limit, window).ServeHTTP(w, r)
		},
	}

	return &HTTPServer{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           handlerMux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		DiceClient: client,
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
	return s.Shutdown()
}

func (s *HTTPServer) Shutdown() error {
	if err := s.DiceClient.Client.Close(); err != nil {
		log.Printf("Failed to close dice client: %v", err)
	}

	return s.httpServer.Shutdown(context.Background())
}

func (s *HTTPServer) HealthCheck(w http.ResponseWriter, request *http.Request) {
	util.JSONResponse(w, http.StatusOK, map[string]string{"message": "Server is running"})
}

func (s *HTTPServer) CliHandler(w http.ResponseWriter, r *http.Request) {
	diceCmds, err := util.ParseHTTPRequest(r)
	if err != nil {
		http.Error(w, "Error parsing HTTP request", http.StatusBadRequest)
		return
	}

	// Check for blacklisted commands
	if containsBlacklistedCommand(strings.ToUpper(diceCmds.Cmd)) {
		http.Error(w, "Error: One or more commands are not allowed", http.StatusForbidden)
		return
	}

	resp := s.DiceClient.ExecuteCommand(diceCmds)
	util.JSONResponse(w, http.StatusOK, resp)
}

func (s *HTTPServer) SearchHandler(w http.ResponseWriter, request *http.Request) {
	util.JSONResponse(w, http.StatusOK, map[string]string{"message": "Search results"})
}

func containsBlacklistedCommand(Cmds string) bool {
	for _, blackListedCmd := range constants.BlacklistedCommands {
		if Cmds == blackListedCmd {
			return true
		}
	}
	return false
}
