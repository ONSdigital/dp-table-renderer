package config

import (
	"time"

	"github.com/ONSdigital/go-ns/log"
	"github.com/kelseyhightower/envconfig"
)

// Config is the filing resource handler config
type Config struct {
	Host               string        `envconfig:"HOST"`
	BindAddr           string        `envconfig:"BIND_ADDR"`
	CORSAllowedOrigins string        `envconfig:"CORS_ALLOWED_ORIGINS"`
	ShutdownTimeout    time.Duration `envconfig:"SHUTDOWN_TIMEOUT"`
}

var cfg *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		Host:               "http://localhost:23100",
		BindAddr:           ":23100",
		CORSAllowedOrigins: "*",
		ShutdownTimeout:    5 * time.Second,
	}

	return cfg, envconfig.Process("", cfg)
}

// Log writes all config properties to log.Debug
func (cfg *Config) Log() {
	log.Debug("Configuration", log.Data{
		"Host":               cfg.Host,
		"BindAddr":           cfg.BindAddr,
		"CORSAllowedOrigins": cfg.CORSAllowedOrigins,
		"ShutdownTimeout":    cfg.ShutdownTimeout,
	})

}
