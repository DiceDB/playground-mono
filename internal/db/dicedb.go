/*
this will be the DiceDB client
*/

package db

import (
	"context"
	"log"

	dice "github.com/dicedb/go-dice"
)

var rdb *dice.Client
var ctx = context.Background()

func InitializeDice(){
	rdb = dice.NewClient(&dice.Options{
		Addr: "localhost:7379",
		Password: "",
		DB: 0,
	})

	err := rdb.Ping(ctx).Err()
	if err != nil{
		log.Fatalf("error connecting to DiceDB: %v", err)
	}
	log.Println("connected to DiceDB")
}

func CloseDiceDB(){
	err := rdb.Close()
	if err!=nil{
		log.Fatalf("error closing DiceDB connection: %v", err)
	}
}