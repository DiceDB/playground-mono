package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DiceAddr string
}

var AppConfig Config

func LoadConfig() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	AppConfig = Config{
		DiceAddr: os.Getenv("DICE_ADDR"),
	}

	return nil
}
