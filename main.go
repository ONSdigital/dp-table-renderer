package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-table-renderer/api"
	"github.com/ONSdigital/dp-table-renderer/config"
	"github.com/ONSdigital/log.go/log"
)

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

func main() {
	log.Namespace = "dp-table-renderer"
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Event(ctx, "unable to run application", log.Error(err), log.FATAL)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "unable to retrieve service configuration", log.FATAL, log.Error(err))
		return err
	}

	log.Event(ctx, "got service configuration", log.INFO, log.Data{"config": cfg})

	// Create healthcheck
	versionInfo, err := healthcheck.NewVersionInfo(BuildTime, GitCommit, Version)
	if err != nil {
		log.Event(ctx, "unable to retrieve health check version info", log.ERROR, log.Error(err))
	}
	healthCheck := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	healthCheck.Start(ctx)

	apiErrors := make(chan error, 1)

	api.CreateRendererAPI(ctx, cfg.BindAddr, cfg.CORSAllowedOrigins, apiErrors, &healthCheck)

	// Gracefully shutdown the application closing any open resources.
	gracefulShutdown := func() error {
		log.Event(ctx, "shutdown with timeout", log.Data{"timeout": cfg.ShutdownTimeout}, log.INFO)
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)

		if err = api.Close(ctx); err != nil {
			log.Event(ctx, "error with graceful shutdown", log.Error(err))
			cancel()
			return err
		}

		cancel()

		log.Event(ctx, "Shutdown complete", log.INFO)
		return nil
	}

	for {
		select {
		case err := <-apiErrors:
			log.Event(ctx, "api error received", log.ERROR, log.Error(err))
			gracefulShutdown()
			return err
		case <-signals:
			log.Event(ctx, "os signal received", log.INFO)
			gracefulShutdown()
			return nil
		}
	}
}
