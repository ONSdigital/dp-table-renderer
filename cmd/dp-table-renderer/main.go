package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-table-renderer/api"
	"github.com/ONSdigital/dp-table-renderer/config"
	"github.com/ONSdigital/go-ns/log"
)

func main() {
	log.Namespace = "dp-table-renderer"

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	cfg.Log()

	apiErrors := make(chan error, 1)

	api.CreateRendererAPI(cfg.Host, cfg.BindAddr, apiErrors)

	// Gracefully shutdown the application closing any open resources.
	gracefulShutdown := func() {
		log.Info(fmt.Sprintf("Shutdown with timeout: %s", cfg.ShutdownTimeout), nil)
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)

		if err = api.Close(ctx); err != nil {
			log.Error(err, nil)
		}

		cancel()

		log.Info("Shutdown complete", nil)
		os.Exit(1)
	}

	for {
		select {
		case err := <-apiErrors:
			log.ErrorC("api error received", err, nil)
			gracefulShutdown()
		case <-signals:
			log.Debug("os signal received", nil)
			gracefulShutdown()
		}
	}
}
