package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"server/internal/cmds"
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

type BlacklistedCommandError struct {
	Command string
}

var blacklistedCommands = map[string]bool{
	"FLUSHALL": true, "FLUSHDB": true, "DUMP": true, "ABORT": true,
	"AUTH": true, "CONFIG": true, "SAVE": true, "BGSAVE": true,
	"BGREWRITEAOF": true, "RESTORE": true, "MULTI": true, "EXEC": true,
	"DISCARD": true, "QWATCH": true, "QUNWATCH": true, "LATENCY": true,
	"CLIENT": true, "SLEEP": true, "PERSIST": true,
}

func ParseHTTPRequest(r *http.Request) (*cmds.CommandRequest, error) {
	command := strings.TrimPrefix(r.URL.Path, "/cli/")
	if command == "" {
		return nil, errors.New("invalid command")
	}
	command = strings.ToUpper(command)

	if blacklistedCommands[command] {
		return nil, &BlacklistedCommandError{Command: command}
	}

	var args []string

	// Extract query parameters
	queryParams := r.URL.Query()
	keyPrefix := queryParams.Get(KeyPrefix)

	if keyPrefix != "" && command == JSONIngest {
		args = append(args, keyPrefix)
	}
	// Step 1: Handle JSON body if present
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		if len(body) > 0 {
			var jsonBody map[string]interface{}
			if err := json.Unmarshal(body, &jsonBody); err != nil {
				return nil, err
			}

			if len(jsonBody) == 0 {
				return nil, fmt.Errorf("empty JSON object")
			}

			// Define keys to exclude and process their values first
			// Update as we support more commands
			var priorityKeys = []string{
				Key,
				Keys,
				Field,
				Path,
				Value,
				Values,
				Seconds,
				User,
				Password,
				KeyValues,
				QwatchQuery,
				Offset,
				Member,
				Members,
			}
			for _, key := range priorityKeys {
				if val, exists := jsonBody[key]; exists {
					if key == Keys {
						for _, v := range val.([]interface{}) {
							args = append(args, fmt.Sprintf("%v", v))
						}
						delete(jsonBody, key)
						continue
					}
					if key == Values {
						for _, v := range val.([]interface{}) {
							args = append(args, fmt.Sprintf("%v", v))
						}
						delete(jsonBody, key)
						continue
					}
					// MultiKey operations
					if key == KeyValues {
						// Handle KeyValues separately
						for k, v := range val.(map[string]interface{}) {
							args = append(args, k, fmt.Sprintf("%v", v))
						}
						delete(jsonBody, key)
						continue
					}
					if key == Members {
						for _, v := range val.([]interface{}) {
							args = append(args, fmt.Sprintf("%v", v))
						}
						delete(jsonBody, key)
						continue
					}
					args = append(args, fmt.Sprintf("%v", val))
					delete(jsonBody, key)
				}
			}

			// Process remaining keys in the JSON body
			for key, val := range jsonBody {
				switch v := val.(type) {
				case string:
					// Handle unary operations like 'nx' where value is "true"
					args = append(args, key)
					if !strings.EqualFold(v, True) {
						args = append(args, v)
					}
				case map[string]interface{}, []interface{}:
					// Marshal nested JSON structures back into a string
					jsonValue, err := json.Marshal(v)
					if err != nil {
						return nil, err
					}
					args = append(args, string(jsonValue))
				default:
					args = append(args, key)
					// Append other types as strings
					value := fmt.Sprintf("%v", v)
					if !strings.EqualFold(value, True) {
						args = append(args, value)
					}
				}
			}
		}
	}

	// Step 2: Return the constructed Redis command
	return &cmds.CommandRequest{
		Cmd:  command,
		Args: args,
	}, nil
}

func (e *BlacklistedCommandError) Error() string {
	return fmt.Sprintf("ERR unknown command '%s'", e.Command)
}

func IsBlacklistedCommand(err error) bool {
	_, ok := err.(*BlacklistedCommandError)
	return ok
}

func JSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
