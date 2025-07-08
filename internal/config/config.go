package config

import (
	"os"
)

type Config struct {
	Port           string
	OllamaHost     string
	MaxConcurrency int
	RequestTimeout int
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		OllamaHost:     getEnv("OLLAMA_HOST", "http://localhost:11434"),
		MaxConcurrency: 50,
		RequestTimeout: 30,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}