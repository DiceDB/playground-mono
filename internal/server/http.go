package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
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
	q := request.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "Missing query parameter 'q' ", http.StatusBadRequest)
		return
	}
	if q == "*" {
		q = ""
	}
	f, err := os.Open("https://github.com/DiceDB/playground-web/blob/master/src/data/command.ts")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	var commands map[string]map[string]string
	err = json.Unmarshal(data, &commands)
	if err != nil {
		log.Fatal(err)
	}
	matchingCommands := []map[string]string{}
	for _, command := range commands {

		title, okTitle := command["title"]
		body, okBody := command["body"]
		if okTitle && okBody {
			if strings.Contains(strings.ToLower(title), q) ||
				strings.Contains(strings.ToLower(body), q) {

				highlightedText := strings.ReplaceAll(title, q, "<b>"+q+"</b>")

				matchingCommands = append(matchingCommands, map[string]string{
					"title":  highlightedText,
					"syntax": command["syntax"],
					"body":   body,
					"url":    command["url"],
				})
			}
		}
	}
	if len(matchingCommands) == 0 {
		util.JSONResponse(w, http.StatusOK, map[string]string{"message": "No search results"})
		return
	}
	response := map[string]interface{}{
		"total":   len(matchingCommands),
		"results": matchingCommands,
	}

	util.JSONResponse(w, http.StatusOK, response)
}
