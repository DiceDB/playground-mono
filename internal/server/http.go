package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"server/config"
	"server/internal/db"
	"server/internal/middleware"
	"server/internal/server/utils"
	util "server/util"

	"github.com/dicedb/dicedb-go"
)

type HTTPServer struct {
	httpServer *http.Server
	DiceClient *db.DiceDB
}

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
	middleware.TrailingSlashMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.ToLower(r.URL.Path)
		cim.rateLimiter(w, r, cim.mux)
	})).ServeHTTP(w, r)
}

func NewHTTPServer(addr string, mux *http.ServeMux, diceDBAdminClient *db.DiceDB, diceClient *db.DiceDB,
	limit int64, window float64) *HTTPServer {
	handlerMux := &HandlerMux{
		mux: mux,
		rateLimiter: func(w http.ResponseWriter, r *http.Request, next http.Handler) {
			middleware.RateLimiter(diceDBAdminClient, next, limit, window).ServeHTTP(w, r)
		},
	}

	return &HTTPServer{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           handlerMux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		DiceClient: diceClient,
	}
}

func (s *HTTPServer) Run(ctx context.Context) error {
	var err error

	go func() {
		slog.Info("starting HTTP server at", slog.String("addr", s.httpServer.Addr))
		if err = s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error: %v", slog.Any("err", err))
		}
	}()

	go func() {
		<-ctx.Done()
		err = s.Shutdown()
		if err != nil {
			slog.Error("Failed to gracefully shutdown HTTP server", slog.Any("err", err))
			return
		}
	}()

	return err
}

func (s *HTTPServer) Shutdown() error {
	if err := s.DiceClient.Client.Close(); err != nil {
		slog.Error("failed to close dicedb client: %v", slog.Any("err", err))
	}

	return s.httpServer.Shutdown(context.Background())
}

func (s *HTTPServer) HealthCheck(w http.ResponseWriter, request *http.Request) {
	nextCleanup, err := s.getNextCleanupTime()
	if err != nil {
		slog.Error("Failed to get the cleanupTime", slog.Any("err", err))
		http.Error(w, errorResponse("internal server error"), http.StatusInternalServerError)
		return
	}

	commandsLeft, err := s.getCommandsLeft()
	if err != nil {
		slog.Error("Failed to get the commands Left", slog.Any("err", err))
		http.Error(w, errorResponse("internal server error"), http.StatusInternalServerError)
		return
	}

	w.Header().Set("x-next-cleanup-time", strconv.FormatInt(nextCleanup, 10))
	w.Header().Set("x-commands-left", strconv.FormatInt(commandsLeft, 10))
	util.JSONResponse(w, http.StatusOK, map[string]string{"message": "server is running"})
}

func (s *HTTPServer) CliHandler(w http.ResponseWriter, r *http.Request) {
	diceCmd, err := util.ParseHTTPRequest(r)
	if err != nil {
		http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
		return
	}

	resp, err := s.DiceClient.ExecuteCommand(diceCmd)
	if err != nil {
		slog.Error("error: failure in executing command", "error", slog.Any("err", err))
		http.Error(w, errorResponse(err.Error()), http.StatusBadRequest)
		return
	}

	respStr, ok := resp.(string)
	if !ok {
		slog.Error("error: response is not a string", "error", slog.Any("err", err))
		http.Error(w, errorResponse("internal server error"), http.StatusInternalServerError)
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

func (s *HTTPServer) getNextCleanupTime() (int64, error) {
	resp := s.DiceClient.Client.Get(context.Background(), utils.LastCronCleanupTimeUnixMs)
	if resp.Err() != nil {
		if errors.Is(resp.Err(), dicedb.Nil) {
			return time.Now().UnixMilli(), nil
		}
		return 0, resp.Err()
	}

	lastCleanupStr := resp.Val()
	if lastCleanupStr == "" {
		return 0, resp.Err()
	}

	lastCleanup, err := strconv.ParseInt(lastCleanupStr, 10, 64)
	if err != nil {
		return 0, err
	}
	lastCleanupTime := time.UnixMilli(lastCleanup)
	nextCleanupTime := lastCleanupTime.Add(15 * time.Minute).UnixMilli()

	return nextCleanupTime, nil
}

func (s *HTTPServer) getCommandsLeft() (int64, error) {
	configValue := config.LoadConfig()
	currentWindow := time.Now().Unix() / int64(configValue.Server.RequestWindowSec)
	key := fmt.Sprintf("request_count:%d", currentWindow)

	val, err := s.DiceClient.Client.Get(context.Background(), key).Result()
	if err != nil {
		if errors.Is(err, dicedb.Nil) {
			return 1000, nil
		}
		return 0, err
	}

	requestCount, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}

	remaining := int64(1000) - requestCount
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}
