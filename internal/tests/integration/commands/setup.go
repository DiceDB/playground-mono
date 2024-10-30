package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"server/config"
	"server/internal/db"
	"server/internal/server"
	"time"
)

type HTTPCommand struct {
	Command string
	Body    []string
}

type CommandExecutor interface {
	FireCommand(cmd string) interface{}
}

type HTTPCommandExecutor struct {
	httpServer *server.HTTPServer
}

type TestCaseResult struct {
	Expected      string
	ErrorExpected bool
}

type TestCase struct {
	Name     string
	Commands []HTTPCommand
	Result   []TestCaseResult
	Delays   []time.Duration
}

func NewHTTPCommandExecutor() (*HTTPCommandExecutor, error) {
	diceClient, err := db.InitDiceClient(config.AppConfig, false)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DiceDB client: %v", err)
	}

	httpServer := &server.HTTPServer{
		DiceClient: diceClient,
	}

	return &HTTPCommandExecutor{
		httpServer,
	}, nil
}

func (hce *HTTPCommandExecutor) FireCommand(httpCommand HTTPCommand) (resp string, err error) {
	body, err := json.Marshal(httpCommand.Body)
	if err != nil {
		return "", fmt.Errorf("error while marshaling reqBody: %v", err)
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "POST", "/shell/exec/"+httpCommand.Command, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("error creating new http request %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(hce.httpServer.CliHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		var cmdErr struct {
			Error string `json:"error"`
		}
		err = json.Unmarshal(rr.Body.Bytes(), &cmdErr)
		if err != nil {
			return "", fmt.Errorf("failed to parse error: %s - %v", rr.Body.String(), err)
		}

		return "", errors.New(cmdErr.Error)
	}

	var cmdResp struct {
		Data string `json:"data"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &cmdResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse command executor response: %s - %v", rr.Body.String(), err)
	}

	return cmdResp.Data, nil
}

func (hce *HTTPCommandExecutor) FlushDB() error {
	flushCmd := HTTPCommand{
		Command: "FLUSHDB",
	}

	_, err := hce.FireCommand(flushCmd)
	if err != nil {
		return fmt.Errorf("error in flushing DB: %v", err)
	}
	return nil
}
