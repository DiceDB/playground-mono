package helpers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"server/internal/cmds"
)

func ParseHTTPRequest(r *http.Request) (*cmds.CommandRequest, error) {
	command := r.PathValue("cmd")
	if command == "" {
		return nil, errors.New("invalid command")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("error reading body: %v", err)
	}

	var commandRequestArgs *cmds.CommandRequestArgs
	err = json.Unmarshal(body, &commandRequestArgs)
	if err != nil {
		log.Fatalf("error unmarshalling body: %v", err)
	}

	return &cmds.CommandRequest{
		Cmd:  command,
		Args: *commandRequestArgs,
	}, nil
}
