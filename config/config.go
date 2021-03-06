package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config is the configuration for this service
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	CORSAllowedOrigins         string        `envconfig:"CORS_ALLOWED_ORIGINS"`
	ShutdownTimeout            time.Duration `envconfig:"SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
}

var cfg *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                   ":23300",
		CORSAllowedOrigins:         "*",
		ShutdownTimeout:            5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
	}

	return cfg, envconfig.Process("", cfg)
}
