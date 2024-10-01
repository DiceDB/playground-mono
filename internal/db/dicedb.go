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

const (
	RespOK = "OK"
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

// ExecuteCommand executes a command based on the input
func (db *DiceDB) ExecuteCommand(command *cmds.CommandRequest) (interface{}, error) {
	switch command.Cmd {
	case "GET":
		if len(command.Args) != 1 {
			return nil, errors.New("invalid args")
		}

		val, err := db.getKey(command.Args[0])
		switch {
		case errors.Is(err, dice.Nil):
			return nil, errors.New("key does not exist")
		case err != nil:
			return nil, fmt.Errorf("get failed %v", err)
		}

		return val, nil

	case "SET":
		if len(command.Args) < 2 {
			return nil, errors.New("key is required")
		}

		err := db.setKey(command.Args[0], command.Args[1])
		if err != nil {
			return nil, errors.New("failed to set key")
		}

		return RespOK, nil

	case "DEL":
		if len(command.Args) == 0 {
			return nil, errors.New("at least one key is required")
		}

		err := db.deleteKeys(command.Args)
		if err != nil {
			return nil, errors.New("failed to delete keys")
		}

		return RespOK, nil

	default:
		return nil, errors.New("unknown command")
	}
}
