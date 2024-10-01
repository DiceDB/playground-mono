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
	DiceAddr       string
	ServerPort     string
	RequestLimit   int      // Field for the request limit
	RequestWindow  float64  // Field for the time window in float64
	AllowedOrigins []string // Field for the allowed origins
}

// LoadConfig loads the application configuration from environment variables or defaults
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found, falling back to system environment variables.")
	}

	return &Config{
		DiceAddr:       getEnv("DICE_ADDR", "localhost:7379"),                                  // Default Dice address
		ServerPort:     getEnv("SERVER_PORT", ":8080"),                                         // Default server port
		RequestLimit:   getEnvInt("REQUEST_LIMIT", 1000),                                       // Default request limit
		RequestWindow:  getEnvFloat64("REQUEST_WINDOW", 60),                                    // Default request window in float64
		AllowedOrigins: getEnvArray("ALLOWED_ORIGINS", []string{"*", "http://localhost:8080"}), // Default allowed origins
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
func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
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

// splitString splits a string by comma and returns a slice of strings
func splitString(s string) []string {
	var array []string
	for _, v := range strings.Split(s, ",") {
		array = append(array, strings.TrimSpace(v))
	}
	return array
}
