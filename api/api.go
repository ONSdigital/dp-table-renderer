package api

import (
	"context"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	"github.com/ONSdigital/dp-table-renderer/config"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var httpServer *dphttp.Server

// RendererAPI manages rendering tables from json
type RendererAPI struct {
	router *mux.Router
}

// CreateRendererAPI manages all the routes configured to the renderer
func CreateRendererAPI(ctx context.Context, bindAddr string, allowedOrigins string, errorChan chan error, hc *healthcheck.HealthCheck) {
	router := mux.NewRouter()
	routes(router, hc)

	cfg, err := config.Get()
	if err != nil {
		log.Error(ctx, "error occurred when getting config", err)
		return
	}

	if cfg.OtelEnabled {
		otelhandler := otelhttp.NewHandler(router, "/")
		httpServer = dphttp.NewServer(bindAddr, otelhandler)
	} else {
		httpServer = dphttp.NewServer(bindAddr, router)
	}

	// Disable this here to allow main to manage graceful shutdown of the entire app.
	httpServer.HandleOSSignals = false

	go func() {
		log.Info(ctx, "starting table renderer")
		if err := httpServer.ListenAndServe(); err != nil {
			log.Error(ctx, "error occurred when running ListenAndServe", err)
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

	cfg, err := config.Get()
	if err != nil {
		log.Error(context.Background(), "error occurred when getting config", err)
		return nil
	}

	var handler http.Handler

	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		if cfg.OtelEnabled {
			handler = otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		} else {
			handler = HttpHandlerTag(pattern, http.HandlerFunc(handlerFunc))
		}
		api.router.Handle(pattern, handler)
	}

	handleFunc("/render/{render_type}", api.renderTable)
	handleFunc("/parse/html", api.parseHTML)

	api.router.StrictSlash(true).Path("/health").HandlerFunc(hc.Handler)

	return &api
}

// Close represents the graceful shutting down of the http server
func Close(ctx context.Context) error {
	if err := httpServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Info(ctx, "graceful shutdown of http server complete")
	return nil
}

func HttpHandlerTag(route string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}
