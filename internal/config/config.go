package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	RailwayClientID     string
	RailwayClientSecret string
	RailwayWorkspaceID  string
	BackendURL          string
	SessionSecret       string
	PublicURL           string
	Port                int
	AuthPrefix          string
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

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	cfg := &Config{
		RailwayClientID:     os.Getenv("RAILWAY_CLIENT_ID"),
		RailwayClientSecret: os.Getenv("RAILWAY_CLIENT_SECRET"),
		RailwayWorkspaceID:  os.Getenv("RAILWAY_WORKSPACE_ID"),
		BackendURL:          os.Getenv("TURNSTILE_BACKEND_URL"),
		SessionSecret:       os.Getenv("TURNSTILE_SESSION_SECRET"),
		PublicURL:           os.Getenv("TURNSTILE_PUBLIC_URL"),
		Port:                port,
		AuthPrefix:          authPrefix,
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
	if c.RailwayWorkspaceID == "" {
		return fmt.Errorf("RAILWAY_WORKSPACE_ID is required")
	}
	if c.BackendURL == "" {
		return fmt.Errorf("TURNSTILE_BACKEND_URL is required")
	}
	if c.SessionSecret == "" {
		return fmt.Errorf("TURNSTILE_SESSION_SECRET is required")
	}
	if len(c.SessionSecret) < 32 {
		return fmt.Errorf("TURNSTILE_SESSION_SECRET must be at least 32 characters")
	}
	if c.PublicURL == "" {
		return fmt.Errorf("TURNSTILE_PUBLIC_URL is required")
	}
	return nil
}
