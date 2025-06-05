package main

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_newHTTPHandler(t *testing.T) {
	tests := []struct {
		name string
		path string
		code int
	}{
		{"health should work", "/healthz", 200},
		{"readiness should work", "/readyz", 200},
		{"random should not be found", "/random", 404},
	}
	srv := httptest.NewServer(newHTTPHandler())
	defer srv.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := http.Get(fmt.Sprintf("%s%s", srv.URL, tt.path))
			if err != nil {
				t.Errorf("calling server failed with: %v", err)
			}

			if got.StatusCode != tt.code {
				t.Errorf("path %s = %v, want %v", tt.path, got, tt.code)
			}
		})
	}
}

func Test_weightedRandom(t *testing.T) {
	codes := []Item{
		{http.StatusOK, 70},
		{http.StatusBadRequest, 20},
		{http.StatusInternalServerError, 10},
	}

	expected := map[int]float64{
		200: 0.7,
		400: 0.2,
		500: 0.1,
	}

	counts := make(map[int]int)
	runs := 100000
	for range runs {
		item := weightedRandom(codes)
		counts[item.Code]++
	}

	margin := 0.02 // 2%
	for name, exp := range expected {
		actual := float64(counts[name]) / float64(runs)
		diff := math.Abs(actual - exp)
		if diff > margin {
			t.Errorf("FAIL: %d is outside the margin of error\n", name)
		}
	}
}
