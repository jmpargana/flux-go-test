package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func OTelMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		handler := otelhttp.WithRouteTag(routePattern, h)
		handler.ServeHTTP(w, r)
	})
}
