package api

import (
	"context"

	"github.com/ONSdigital/dp-table-renderer/health"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

var httpServer *server.Server

// RendererAPI manages rendering tables from json
type RendererAPI struct {
	host   string
	router *mux.Router
}

// CreateRendererAPI manages all the routes configured to the renderer
func CreateRendererAPI(host, bindAddr string, errorChan chan error) {
	router := mux.NewRouter()
	routes(host, router)

	httpServer = server.New(bindAddr, router)
	// Disable this here to allow main to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Debug("Starting table renderer...", nil)
		if err := httpServer.ListenAndServe(); err != nil {
			log.ErrorC("Table renderer http server returned error", err, nil)
			errorChan <- err
		}
	}()
}

// routes contain all endpoints for the renderer
func routes(host string, router *mux.Router) *RendererAPI {
	api := RendererAPI{host: host, router: router}

	router.Path("/healthcheck").Methods("GET").HandlerFunc(health.EmptyHealthcheck)

	api.router.HandleFunc("/render/{render_type}", api.renderTable).Methods("POST")
	return &api
}

// Close represents the graceful shutting down of the http server
func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Info("graceful shutdown of http server complete", nil)
	return nil
}
