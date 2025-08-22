package config

import (
	"os"
	"strconv"
)

type Config struct {
	Environment string
	Port        string
	DatabaseURL string
	JWTSecret   string
	LogLevel    string
	LogToFile   bool
}

func Load() *Config {
	logToFile, _ := strconv.ParseBool(getEnv("LOG_TO_FILE", "true"))
	
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "devops.db"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		LogToFile:   logToFile,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
