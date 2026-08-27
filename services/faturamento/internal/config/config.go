package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	EstoqueURL     string
	EstoqueTimeout time.Duration
}

func Load() Config {
	return Config{
		Port:           getEnv("PORT", "8082"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5435"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBName:         getEnv("DB_NAME", "faturamento_db"),
		EstoqueURL:     getEnv("ESTOQUE_SERVICE_URL", "http://localhost:8081"),
		EstoqueTimeout: getEnvDuration("ESTOQUE_TIMEOUT_MS", 2000) * time.Millisecond,
	}
}

func getEnvDuration(key string, fallback int) time.Duration {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return time.Duration(parsed)
		}
	}
	return time.Duration(fallback)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
