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
	"server/util/cmds"
	"strings"
	"time"

	dicedb "github.com/dicedb/go-dice"
)

const RespNil = "(nil)"

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

func InitDiceClient(DiceDBAddr string) (*DiceDB, error) {
	diceClient := dicedb.NewClient(&dicedb.Options{
		Addr:        DiceDBAddr,
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
		return RespNil, nil
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
		return renderListResponse(v)
	case int64:
		return fmt.Sprintf("%v", v), nil
	case nil:
		return RespNil, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func renderListResponse(items []interface{}) (string, error) {
	if len(items)%2 != 0 {
		return "", fmt.Errorf("(error) invalid result format")
	}

	var builder strings.Builder
	for i := 0; i < len(items); i += 2 {
		field, ok1 := items[i].(string)
		value, ok2 := items[i+1].(string)

		// Check if both field and value are valid strings
		if !ok1 || !ok2 {
			return "", fmt.Errorf("(error) invalid result type")
		}

		// Append the formatted field and value
		_, err := fmt.Fprintf(&builder, "%d) \"%s\"\n%d) \"%s\"\n", i+1, field, i+2, value)
		if err != nil {
			return "", err
		}
	}

	return builder.String(), nil
}
