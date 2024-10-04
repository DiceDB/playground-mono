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
	cmds "server/util/cmds"
	"strings"
)

const (
	Key         = "key"
	Keys        = "keys"
	KeyPrefix   = "key_prefix"
	Field       = "field"
	Path        = "path"
	Value       = "value"
	Values      = "values"
	User        = "user"
	Password    = "password"
	Seconds     = "seconds"
	KeyValues   = "key_values"
	True        = "true"
	QwatchQuery = "query"
	Offset      = "offset"
	Member      = "member"
	Members     = "members"

	JSONIngest string = "JSON.INGEST"
)

var priorityKeys = []string{
	Key, Keys, Field, Path, Value, Values, Seconds, User, Password, KeyValues, QwatchQuery, Offset, Member, Members,
}

// ParseHTTPRequest parses an incoming HTTP request and converts it into a CommandRequest for Redis commands
func ParseHTTPRequest(r *http.Request) (*cmds.CommandRequest, error) {
	command := extractCommand(r.URL.Path)
	if command == "" {
		return nil, errors.New("invalid command")
	}

	args, err := extractArgsFromRequest(r, command)
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

func extractArgsFromRequest(r *http.Request, command string) ([]string, error) {
	var args []string
	queryParams := r.URL.Query()
	keyPrefix := queryParams.Get(KeyPrefix)

	if keyPrefix != "" && command == JSONIngest {
		args = append(args, keyPrefix)
	}

	if r.Body != nil {
		bodyArgs, err := parseRequestBody(r.Body)
		if err != nil {
			return nil, err
		}
		args = append(args, bodyArgs...)
	}

	return args, nil
}

func parseRequestBody(body io.ReadCloser) ([]string, error) {
	var args []string
	bodyContent, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	if len(bodyContent) == 0 {
		return args, nil
	}

	var jsonBody map[string]interface{}
	if err := json.Unmarshal(bodyContent, &jsonBody); err != nil {
		return nil, err
	}

	if len(jsonBody) == 0 {
		return nil, fmt.Errorf("empty JSON object")
	}

	args = append(args, extractPriorityArgs(jsonBody)...)
	args = append(args, extractRemainingArgs(jsonBody)...)

	return args, nil
}

func extractPriorityArgs(jsonBody map[string]interface{}) []string {
	var args []string
	for _, key := range priorityKeys {
		if val, exists := jsonBody[key]; exists {
			switch key {
			case Keys, Values, Members:
				args = append(args, convertListToStrings(val.([]interface{}))...)
			case KeyValues:
				args = append(args, convertMapToStrings(val.(map[string]interface{}))...)
			default:
				args = append(args, fmt.Sprintf("%v", val))
			}
			delete(jsonBody, key)
		}
	}
	return args
}

func extractRemainingArgs(jsonBody map[string]interface{}) []string {
	var args []string
	for key, val := range jsonBody {
		switch v := val.(type) {
		case string:
			args = append(args, key)
			if !strings.EqualFold(v, True) {
				args = append(args, v)
			}
		case map[string]interface{}, []interface{}:
			jsonValue, _ := json.Marshal(v)
			args = append(args, string(jsonValue))
		default:
			args = append(args, key, fmt.Sprintf("%v", v))
		}
	}
	return args
}

func convertListToStrings(list []interface{}) []string {
	var result []string
	for _, v := range list {
		result = append(result, fmt.Sprintf("%v", v))
	}
	return result
}

func convertMapToStrings(m map[string]interface{}) []string {
	var result []string
	for k, v := range m {
		result = append(result, k, fmt.Sprintf("%v", v))
	}
	return result
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
