package api

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"net/http"
)

var httpServer *server.Server

// RendererAPI manages rendering tables from json
type RendererAPI struct {
	router *mux.Router
}

// CreateRendererAPI manages all the routes configured to the renderer
func CreateRendererAPI(bindAddr string, allowedOrigins string, errorChan chan error, hc *healthcheck.HealthCheck) {
	router := mux.NewRouter()
	routes(router, hc)

	httpServer = server.New(bindAddr, createCORSHandler(allowedOrigins, router))
	// Disable this here to allow main to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Debug("Starting table renderer...", nil)
		if err := httpServer.ListenAndServe(); err != nil {
			log.ErrorC("Main", err, log.Data{"MethodInError": "httpServer.ListenAndServe()"})
			errorChan <- err
		}
	}()
}

// createCORSHandler wraps the router in a CORS handler that responds to OPTIONS requests and returns the headers necessary to allow CORS-enabled clients to work
func createCORSHandler(allowedOrigins string, router *mux.Router) http.Handler {
	headersOk := handlers.AllowedHeaders([]string{"Accept", "Content-Type", "Access-Control-Allow-Origin", "Access-Control-Allow-Methods", "X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{allowedOrigins})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"})

	return handlers.CORS(originsOk, headersOk, methodsOk)(router)
}

// routes contain all endpoints for the renderer
func routes(router *mux.Router, hc *healthcheck.HealthCheck) *RendererAPI {
	api := RendererAPI{router: router}

	api.router.StrictSlash(true).Path("/health").HandlerFunc(hc.Handler)
	api.router.HandleFunc("/render/{render_type}", api.renderTable).Methods("POST")
	api.router.HandleFunc("/parse/html", api.parseHTML).Methods("POST")
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
