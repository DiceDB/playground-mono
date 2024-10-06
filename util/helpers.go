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

// Map of blocklisted commands
var blocklistedCommands = map[string]bool{
	"FLUSHALL":     true,
	"FLUSHDB":      true,
	"DUMP":         true,
	"ABORT":        true,
	"AUTH":         true,
	"CONFIG":       true,
	"SAVE":         true,
	"BGSAVE":       true,
	"BGREWRITEAOF": true,
	"RESTORE":      true,
	"MULTI":        true,
	"EXEC":         true,
	"DISCARD":      true,
	"QWATCH":       true,
	"QUNWATCH":     true,
	"LATENCY":      true,
	"CLIENT":       true,
	"SLEEP":        true,
	"PERSIST":      true,
}

// BlockListedCommand checks if a command is blocklisted
func BlockListedCommand(cmd string) error {
	if _, exists := blocklistedCommands[strings.ToUpper(cmd)]; exists {
		return errors.New("ERR unknown command '" + cmd + "'")
	}
	return nil
}

// ParseHTTPRequest parses an incoming HTTP request and converts it into a CommandRequest for Redis commands
func ParseHTTPRequest(r *http.Request) (*cmds.CommandRequest, error) {
	command := extractCommand(r.URL.Path)
	if command == "" {
		return nil, errors.New("invalid command")
	}

	// Check if the command is blocklisted
	if err := BlockListedCommand(command); err != nil {
		return nil, err
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

	if len(jsonBody) == 0 {
		return nil, fmt.Errorf("empty JSON object")
	}

	for _, val := range jsonBody {
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("invalid input")
		}

		args = append(args, s)
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
