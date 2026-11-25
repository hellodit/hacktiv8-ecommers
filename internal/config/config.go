package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Port      string
	DSN       string
	JWTSecret []byte
	JWTTTL    time.Duration
}

func Load() (*Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	ttl, err := time.ParseDuration(getEnv("JWT_TTL", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_TTL: %w", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "ecommers"),
		getEnv("DB_SSLMODE", "disable"),
	)

	return &Config{
		Port:      getEnv("PORT", "8080"),
		DSN:       dsn,
		JWTSecret: []byte(secret),
		JWTTTL:    ttl,
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
