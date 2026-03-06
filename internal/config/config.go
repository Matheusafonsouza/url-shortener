package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	DatabaseURL     string
	BaseURL         string
	DefaultTTL      time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	defaultTTL, err := parseDurationEnv("DEFAULT_TTL", "720h")
	if err != nil {
		return Config{}, err
	}

	readTimeout, err := parseDurationEnv("READ_TIMEOUT", "5s")
	if err != nil {
		return Config{}, err
	}

	writeTimeout, err := parseDurationEnv("WRITE_TIMEOUT", "10s")
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := parseDurationEnv("SHUTDOWN_TIMEOUT", "10s")
	if err != nil {
		return Config{}, err
	}

	port := envOrDefault("PORT", "8080")
	if _, err := strconv.Atoi(port); err != nil {
		return Config{}, fmt.Errorf("invalid PORT: %w", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	baseURL := envOrDefault("BASE_URL", "http://localhost:"+port)

	return Config{
		Port:            port,
		DatabaseURL:     databaseURL,
		BaseURL:         baseURL,
		DefaultTTL:      defaultTTL,
		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}

func envOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func parseDurationEnv(key, fallback string) (time.Duration, error) {
	raw := envOrDefault(key, fallback)
	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}

	return duration, nil
}
