package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"server/internal/middleware"
	db "server/internal/tests/dbmocks"
	"server/util/cmds"
	"strings"
)

// ParseHTTPRequest parses an incoming HTTP request and converts it into a CommandRequest for Redis commands
func ParseHTTPRequest(r *http.Request) (*cmds.CommandRequest, error) {
	command := extractCommand(r.URL.Path)
	if command == "" {
		return nil, errors.New("invalid command")
	}

	args, err := newExtractor(r)
	if err != nil {
		return nil, err
	}

	return &cmds.CommandRequest{
		Cmd:  command,
		Args: args,
	}, nil
}

func extractCommand(path string) string {
	command := strings.TrimPrefix(path, "/shell/exec/")
	return strings.ToUpper(command)
}

func newExtractor(r *http.Request) ([]string, error) {
	var args []string
	bodyContent, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if len(bodyContent) == 0 {
		return args, nil
	}

	var jsonBody []interface{}
	if err := json.Unmarshal(bodyContent, &jsonBody); err != nil {
		return nil, err
	}
	for _, val := range jsonBody {
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("invalid input")
		}
		if strings.TrimSpace(s) != "" {
			args = append(args, s)
		}
	}

	return args, nil
}

// JSONResponse sends a JSON response to the client
func JSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// MockHandler is a basic mock handler for testing
func MockHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		slog.Error("Failed to write response: %v", slog.Any("err", err))
	}
}

// SetupRateLimiter sets up a rate limiter for testing purposes
func SetupRateLimiter(limit int64, window float64) (*httptest.ResponseRecorder, *http.Request, http.Handler) {
	mockClient := db.NewDiceDBMock()

	r := httptest.NewRequest("GET", "/shell/exec/get", http.NoBody)
	w := httptest.NewRecorder()

	rateLimiter := middleware.MockRateLimiter(mockClient, http.HandlerFunc(MockHandler), limit, window)

	return w, r, rateLimiter
}
