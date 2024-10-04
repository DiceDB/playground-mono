package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"server/internal/middleware"
	"strings"
	"sync"
	"time"

	"server/internal/db"
	util "server/util"
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

type HTTPResponse struct {
	Data interface{} `json:"data"`
}

type HTTPErrorResponse struct {
	Error interface{} `json:"error"`
}

func errorResponse(response string) string {
	errorMessage := map[string]string{"error": response}
	jsonResponse, err := json.Marshal(errorMessage)
	if err != nil {
		slog.Error("Error marshaling response: %v", slog.Any("err", err))
		return `{"error": "internal server error"}`
	}
	return string(jsonResponse)
}

func (cim *HandlerMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Convert the path to lowercase before passing to the underlying mux.
	middleware.TrailingSlashMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.ToLower(r.URL.Path)
		// Apply rate limiter
		cim.rateLimiter(w, r, cim.mux)
	})).ServeHTTP(w, r)
}

func NewHTTPServer(addr string, mux *http.ServeMux, client *db.DiceDB, limit int64, window float64) *HTTPServer {
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
		slog.Info("starting server at", slog.String("addr", s.httpServer.Addr))
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error: %v", slog.Any("err", err))
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server...")
	return s.Shutdown()
}

func (s *HTTPServer) Shutdown() error {
	if err := s.DiceClient.Client.Close(); err != nil {
		slog.Error("failed to close dicedb client: %v", slog.Any("err", err))
	}

	return s.httpServer.Shutdown(context.Background())
}

func (s *HTTPServer) HealthCheck(w http.ResponseWriter, request *http.Request) {
	util.JSONResponse(w, http.StatusOK, map[string]string{"message": "server is running"})
}

func (s *HTTPServer) CliHandler(w http.ResponseWriter, r *http.Request) {
	diceCmd, err := util.ParseHTTPRequest(r)
	if err != nil {
		http.Error(w, errorResponse("error parsing http request"), http.StatusBadRequest)
		return
	}

	resp, err := s.DiceClient.ExecuteCommand(diceCmd)
	if err != nil {
		http.Error(w, errorResponse("error executing command"), http.StatusBadRequest)
		return
	}

	respStr, ok := resp.(string)
	if !ok {
		slog.Error("error: response is not a string", "error", slog.Any("err", err))
		http.Error(w, errorResponse("internal Server Error"), http.StatusInternalServerError)
		return
	}

	httpResponse := HTTPResponse{Data: respStr}
	responseJSON, err := json.Marshal(httpResponse)
	if err != nil {
		slog.Error("error marshaling response to json", "error", slog.Any("err", err))
		http.Error(w, errorResponse("internal server error"), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(responseJSON)
	if err != nil {
		http.Error(w, errorResponse("internal server error"), http.StatusInternalServerError)
		return
	}
}

func (s *HTTPServer) SearchHandler(w http.ResponseWriter, request *http.Request) {
	util.JSONResponse(w, http.StatusOK, map[string]string{"message": "search results"})
}