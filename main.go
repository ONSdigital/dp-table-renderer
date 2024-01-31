package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dpotelgo "github.com/ONSdigital/dp-otel-go"
	"github.com/ONSdigital/dp-table-renderer/api"
	"github.com/ONSdigital/dp-table-renderer/config"
	"github.com/ONSdigital/log.go/v2/log"
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

	if err := run(ctx); err != nil && err != http.ErrServerClosed {
		log.Fatal(ctx, "unable to run application", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "unable to retrieve service configuration", err)
		return err
	}

	if cfg.OtelEnabled {
		//Set up OpenTelemetry
		otelConfig := dpotelgo.Config{
			OtelServiceName:          cfg.OTServiceName,
			OtelExporterOtlpEndpoint: cfg.OTExporterOTLPEndpoint,
			OtelBatchTimeout:         cfg.OTBatchTimeout,
		}

		otelShutdown, err := dpotelgo.SetupOTelSDK(ctx, otelConfig)

		if err != nil {
			log.Error(ctx, "error setting up OpenTelemetry - hint: ensure OTEL_EXPORTER_OTLP_ENDPOINT is set", err)
		}
		// Handle shutdown properly so nothing leaks.
		defer func() {
			err = errors.Join(err, otelShutdown(context.Background()))
		}()

		otelShutdown(ctx)
	}

	log.Info(ctx, "got service configuration", log.Data{"config": cfg})

	// Create healthcheck
	versionInfo, err := healthcheck.NewVersionInfo(BuildTime, GitCommit, Version)
	if err != nil {
		log.Error(ctx, "unable to retrieve health check version info", err)
	}
	healthCheck := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	healthCheck.Start(ctx)

	apiErrors := make(chan error, 1)

	api.CreateRendererAPI(ctx, cfg.BindAddr, cfg.CORSAllowedOrigins, apiErrors, &healthCheck)

	// Gracefully shutdown the application closing any open resources.
	gracefulShutdown := func() error {
		log.Info(ctx, "shutdown with timeout", log.Data{"timeout": cfg.ShutdownTimeout})
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err = api.Close(ctx); err != nil {
			log.Error(ctx, "error with graceful shutdown", err)
			cancel()
			return err
		}

		log.Info(ctx, "Shutdown complete")
		return nil
	}

	for {
		select {
		case err := <-apiErrors:
			log.Error(ctx, "api error received", err)
			gracefulShutdown()
			return err
		case <-signals:
			log.Info(ctx, "os signal received")
			gracefulShutdown()
			return nil
		}
	}
}
