package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	DiceDB struct {
		Addr     string // Field for the Dice address
		Username string // Field for the username
		Password string // Field for the password
	}
	Server struct {
		Port               string // Field for the server port
		IsTestEnv          bool
		RequestLimitPerMin int64    // Field for the request limit
		RequestWindowSec   float64  // Field for the time window in float64
		AllowedOrigins     []string // Field for the allowed origins
	}
}

// LoadConfig loads the application configuration from environment variables or defaults
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found, falling back to system environment variables.")
	}

	return &Config{
		DiceDB: struct {
			Addr     string
			Username string
			Password string
		}{
			Addr:     getEnv("DICEDB_ADDR", "localhost:7379"), // Default Dice address
			Username: getEnv("DICEDB_USERNAME", "dice"),       // Default username
			Password: getEnv("DICEDB_PASSWORD", ""),           // Default password
		},
		Server: struct {
			Port               string
			IsTestEnv          bool
			RequestLimitPerMin int64
			RequestWindowSec   float64
			AllowedOrigins     []string
		}{
			Port:               getEnv("SERVER_PORT", ":8080"),
			IsTestEnv:          getEnvBool("IS_TEST_ENVIRONMENT", false),                          // Default server port
			RequestLimitPerMin: getEnvInt("REQUEST_LIMIT_PER_MIN", 1000),                          // Default request limit
			RequestWindowSec:   getEnvFloat64("REQUEST_WINDOW_SEC", 60),                           // Default request window in float64
			AllowedOrigins:     getEnvArray("ALLOWED_ORIGINS", []string{"http://localhost:3000"}), // Default allowed origins
		},
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvInt retrieves an environment variable as an integer or returns a default value
func getEnvInt(key string, fallback int) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return int64(intValue)
		}
	}
	return int64(fallback)
}

// added for miliseconds request window controls
func getEnvFloat64(key string, fallback float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return fallback
}

func getEnvArray(key string, fallback []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		if arrayValue := splitString(value); len(arrayValue) > 0 {
			return arrayValue
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

// splitString splits a string by comma and returns a slice of strings
func splitString(s string) []string {
	var array []string
	for _, v := range strings.Split(s, ",") {
		array = append(array, strings.TrimSpace(v))
	}
	return array
}
