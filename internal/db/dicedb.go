/*
this will be the DiceDB client
*/

package db

import (
	"context"
	"log"
	"server/internal/cmds"

	dice "github.com/dicedb/go-dice"
)

var rdb *dice.Client
var ctx = context.Background()

func CloseDiceDB(){
	err := rdb.Close()
	if err!=nil{
		log.Fatalf("error closing DiceDB connection: %v", err)
	}
}

func errorResponse(response string) map[string]string {
	return map[string]string{"error": response}
}

func ExecuteCommand(command *cmds.CommandRequest) (interface{}){
	switch command.Cmd{
	case "get":
		if command.Args.Key == ""{
			return errorResponse("key is required")
		}
		val, err := getKey(command.Args.Key)
		if err!=nil{
			return errorResponse("error running get command")
		}
		return map[string]string{"value": val}

	case "set":
		if command.Args.Key == "" || command.Args.Value == ""{
			return errorResponse("key and value are required")
		}
		err := setKey(command.Args.Key, command.Args.Value)
		if err!=nil{
			return errorResponse("failed to set key")
		}
		return map[string]string{"result": "OK"}

	case "del":
		if len(command.Args.Keys) == 0{
			return map[string]string{"error": "atleast one key is required"}
		}
		err := deleteKeys(command.Args.Keys)
		if err!=nil{
			return map[string]string{"error": "failed to delete keys"}
		}

		return map[string]string{"result": "OK"}
	default:
		return errorResponse("unknown command")
	}
}