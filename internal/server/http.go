//modified http

package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"server/internal/db"
	"server/internal/middleware"
	util "server/util"
)

type HTTPServer struct {
	httpServer *http.Server
	DiceClient *db.DiceDB
}

type HandlerMux struct {
	mux         *http.ServeMux
	rateLimiter func(http.ResponseWriter, *http.Request, http.Handler)
}


func (cim *HandlerMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	middleware.TrailingSlashMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.ToLower(r.URL.Path)
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
	util.HttpResponseJSON(w, http.StatusOK, map[string]string{"message": "Server is running"})
}

func (s *HTTPServer) CliHandler(w http.ResponseWriter, r *http.Request) {
	diceCmd, err := util.ParseHTTPRequest(r)
	if err != nil {
		util.HttpResponseException(w,http.StatusBadRequest,err)
		return
	}

	resp, err := s.DiceClient.ExecuteCommand(diceCmd)
	if err != nil {
		slog.Error("error: failure in executing command", "error", slog.Any("err", err))
		util.HttpResponseException(w,http.StatusBadRequest,err)
		return
	}
	util.HttpResponseJSON(w, http.StatusOK, resp)
}

func (s *HTTPServer) SearchHandler(w http.ResponseWriter, request *http.Request) {
	util.JSONResponse(w, http.StatusOK, map[string]string{"message": "search results"})
}
