package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Workers   int
	QueueSize int
	Port      string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Workers:   getEnvInt("WORKERS", 3),
		QueueSize: getEnvInt("QUEUE_SIZE", 100),
		Port:      getEnvString("PORT", "8080"),
	}
}

// getEnvInt gets an environment variable as an integer with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvString gets an environment variable as a string with a default value
func getEnvString(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
