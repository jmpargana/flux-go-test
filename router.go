package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func newHTTPHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(OTelMiddleware)
	r.Use(middleware.Logger)
	r.Get("/", rootHandler)
	r.Get("/healthz", probeHandler)
	r.Get("/readyz", probeHandler)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	})
	return otelhttp.NewHandler(r, "/")
}
