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
	"server/util/cmds"
	"time"

	dicedb "github.com/dicedb/go-dice"
)

type DiceDB struct {
	Client *dicedb.Client
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
	diceClient := dicedb.NewClient(&dicedb.Options{
		Addr:        configValue.DiceDBAddr,
		DialTimeout: 10 * time.Second,
		MaxRetries:  10,
	})

	// Ping the dicedb client to verify the connection
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
	args := make([]interface{}, 0, len(command.Args)+1)
	args = append(args, command.Cmd)
	for _, arg := range command.Args {
		args = append(args, arg)
	}

	res, err := db.Client.Do(db.Ctx, args...).Result()
	if errors.Is(err, dicedb.Nil) {
		return nil, errors.New("(nil)")
	}

	if err != nil {
		return nil, fmt.Errorf("(error) %v", err)
	}

	// Print the result based on its type
	switch v := res.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case []interface{}:
	case int64:
		return fmt.Sprintf("%v", v), nil
	case nil:
		return "(nil)", nil
	default:
		return fmt.Sprintf("%v", v), nil
	}

	return nil, fmt.Errorf("(error) invalid result type")
}
