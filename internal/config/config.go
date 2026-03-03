package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	RailwayClientID     string
	RailwayClientSecret string
	RailwayProjectID    string
	BackendURL          string
	PublicURL           string
	Port                int
	AuthPrefix          string
	LogLevel            string
	ProxyMaxRetries     int
	ProxyRetryDelay     time.Duration
}

func Load() (*Config, error) {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}

	authPrefix := os.Getenv("TURNSTILE_AUTH_PREFIX")
	if authPrefix == "" {
		authPrefix = "/_turnstile"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	maxRetriesStr := os.Getenv("TURNSTILE_PROXY_MAX_RETRIES")
	if maxRetriesStr == "" {
		maxRetriesStr = "3"
	}

	retryDelayStr := os.Getenv("TURNSTILE_PROXY_RETRY_DELAY")
	if retryDelayStr == "" {
		retryDelayStr = "1s"
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	maxRetries, err := strconv.Atoi(maxRetriesStr)
	if err != nil {
		return nil, fmt.Errorf("invalid TURNSTILE_PROXY_MAX_RETRIES: %w", err)
	}

	retryDelay, err := time.ParseDuration(retryDelayStr)
	if err != nil {
		return nil, fmt.Errorf("invalid TURNSTILE_PROXY_RETRY_DELAY: %w", err)
	}

	cfg := &Config{
		RailwayClientID:     os.Getenv("RAILWAY_CLIENT_ID"),
		RailwayClientSecret: os.Getenv("RAILWAY_CLIENT_SECRET"),
		RailwayProjectID:    os.Getenv("RAILWAY_PROJECT_ID"),
		BackendURL:          os.Getenv("TURNSTILE_BACKEND_URL"),
		PublicURL:           os.Getenv("TURNSTILE_PUBLIC_URL"),
		Port:                port,
		AuthPrefix:          authPrefix,
		LogLevel:            logLevel,
		ProxyMaxRetries:     maxRetries,
		ProxyRetryDelay:     retryDelay,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.RailwayClientID == "" {
		return fmt.Errorf("RAILWAY_CLIENT_ID is required")
	}
	if c.RailwayClientSecret == "" {
		return fmt.Errorf("RAILWAY_CLIENT_SECRET is required")
	}
	if c.RailwayProjectID == "" {
		return fmt.Errorf("RAILWAY_PROJECT_ID is required")
	}
	if c.BackendURL == "" {
		return fmt.Errorf("TURNSTILE_BACKEND_URL is required")
	}
	if c.PublicURL == "" {
		return fmt.Errorf("TURNSTILE_PUBLIC_URL is required")
	}
	return nil
}
