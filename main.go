package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Item struct {
	Code   int
	Weight int
}

func weightedRandom(items []Item) Item {
	totalWeight := 0
	for _, item := range items {
		totalWeight += item.Weight
	}

	// Generate a random number in the range [0, totalWeight)
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(totalWeight)

	// Pick the item based on the weight
	for _, item := range items {
		if r < item.Weight {
			return item
		}
		r -= item.Weight
	}
	return Item{}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "root")
	defer span.End()
	codes := []Item{
		{http.StatusOK, 70},
		{http.StatusBadRequest, 20},
		{http.StatusInternalServerError, 10},
	}
	code := weightedRandom(codes)
	waitFor, err := time.ParseDuration(fmt.Sprintf("%dms", rand.Intn(2000)))
	if err != nil {
		logger.ErrorContext(ctx, "WaitFor failed: %v\n", slog.Any("error", err))
	}
	time.Sleep(waitFor)

	msg := fmt.Sprintf("%dms -> %d\n", waitFor.Milliseconds(), code.Code)
	logger.InfoContext(ctx, msg, slog.Int("result", code.Code))
	w.WriteHeader(code.Code)
	if _, err := io.WriteString(w, msg); err != nil {
		logger.ErrorContext(ctx, "Write failed: %v\n", slog.Any("error", err))
	}
}

func probeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	srv := &http.Server{
		Addr:         ":8080",
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	select {
	case err = <-srvErr:
		return
	case <-ctx.Done():
		stop()
	}

	err = srv.Shutdown(context.Background())
	return
}

func OTelMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		handler := otelhttp.WithRouteTag(routePattern, h)
		handler.ServeHTTP(w, r)
	})
}

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
