package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	RedisAddr     string
	ServerPort    string
	RequestLimit  int // Field for the request limit
	RequestWindow int // Field for the time window in seconds
}

// LoadConfig loads the application configuration from environment variables or defaults
func LoadConfig() *Config {
	return &Config{
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:7379"), // Default Redis address
		ServerPort:    getEnv("SERVER_PORT", ":8080"),         // Default server port
		RequestLimit:  getEnvInt("REQUEST_LIMIT", 1000),       // Default request limit
		RequestWindow: getEnvInt("REQUEST_WINDOW", 60),        // Default request window in seconds
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
