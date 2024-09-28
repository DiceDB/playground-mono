/*
this will be the DiceDB client
*/

package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"server/config"
	"server/internal/cmds"
	"time"

	dice "github.com/dicedb/go-dice"
)

type DiceDB struct {
	Client *dice.Client
	Ctx    context.Context
}

func (db *DiceDB) CloseDiceDB() {
	err := db.Client.Close()
	if err != nil {
		slog.Error("error closing DiceDB connection",
			slog.Any("error", err))
		os.Exit(1)
	}
}

func InitDiceClient(configValue *config.Config) (*DiceDB, error) {
	diceClient := dice.NewClient(&dice.Options{
		Addr:        configValue.DiceAddr,
		DialTimeout: 10 * time.Second,
		MaxRetries:  10,
	})

	// Ping the dice client to verify the connection
	err := diceClient.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}

	return &DiceDB{
		Client: diceClient,
		Ctx:    context.Background(),
	}, nil
}

func errorResponse(response string) map[string]string {
	return map[string]string{"error": response}
}

// ExecuteCommand executes a command based on the input
func (db *DiceDB) ExecuteCommand(command *cmds.CommandRequest) interface{} {
	switch command.Cmd {
	case "get":
		if command.Args.Key == "" {
			return errorResponse("key is required")
		}

		val, err := db.getKey(command.Args.Key)
		switch {
		case errors.Is(err, dice.Nil):
			return errorResponse("key does not exist")
		case err != nil:
			return errorResponse(fmt.Sprintf("Get failed %v", err))
		}

		return map[string]string{"value": val}

	case "set":
		if command.Args.Key == "" || command.Args.Value == "" {
			return errorResponse("key and value are required")
		}
		err := db.setKey(command.Args.Key, command.Args.Value)
		if err != nil {
			return errorResponse("failed to set key")
		}
		return map[string]string{"result": "OK"}

	case "del":
		if len(command.Args.Keys) == 0 {
			return errorResponse("at least one key is required")
		}
		err := db.deleteKeys(command.Args.Keys)
		if err != nil {
			return errorResponse("failed to delete keys")
		}

		return map[string]string{"result": "OK"}

	default:
		return errorResponse("unknown command")
	}
}
