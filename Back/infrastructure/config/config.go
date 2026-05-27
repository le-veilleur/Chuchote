package config

import (
	"os"
)

type Config struct {
	Port      string
	JWTSecret string
	DatabaseDSN string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "change-me-in-production"),
		DatabaseDSN: getEnv("DATABASE_DSN", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
