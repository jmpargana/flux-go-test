package main

import (
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"time"
)

type item struct {
	Code   int
	Weight int
}

func weightedRandom(items []item) item {
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
	return item{}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "root")
	defer span.End()
	codes := []item{
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
