package db

import (
	"context"
	"log"
	"server/config"

	dice "github.com/dicedb/go-dice"
)

var rdb *dice.Client
var ctx = context.Background()

func InitializeDiceDB() {
	rdb = dice.NewClient(&dice.Options{
		Addr:     config.AppConfig.DiceAddr,
		Password: "",
		DB:       0,
	})
	err := rdb.Ping(ctx).Err()
	if err != nil {
		log.Fatalf("Could not connect to DiceDB: %v", err)
	}
	log.Println("Connected to DiceDB")
}

func SetKey(key, value string) error {
	err := rdb.Set(ctx, key, value, 0).Err()
	return err
}

func GetKey(key string) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	return val, err
}

func CloseDiceDB() {
	if err := rdb.Close(); err != nil {
		log.Fatalf("Error closing DiceDB connection: %v", err)
	}
}
