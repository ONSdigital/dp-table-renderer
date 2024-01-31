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
	OTExporterOTLPEndpoint     string        `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTServiceName              string        `envconfig:"OTEL_SERVICE_NAME"`
	OTBatchTimeout             time.Duration `envconfig:"OTEL_BATCH_TIMEOUT"`
	OtelEnabled                bool          `envconfig:"OTEL_ENABLED"`
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
		OTExporterOTLPEndpoint:     "localhost:4317",
		OTServiceName:              "dp-table-renderer",
		OTBatchTimeout:             5 * time.Second,
		OtelEnabled:                false,
	}

	return cfg, envconfig.Process("", cfg)
}
