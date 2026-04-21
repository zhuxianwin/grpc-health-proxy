package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fixedChecker always returns the given status.
type fixedChecker struct{ status Status }

func (f *fixedChecker) Check(_ context.Context, _ string) Result {
	return Result{Status: f.status}
}

func TestTrend_UnknownBelowMinSample(t *testing.T) {
	tc := NewTrendChecker(&fixedChecker{StatusHealthy}, TrendConfig{Window: 10, MinSample: 5})
	res := tc.Check(context.Background(), "svc")
	if res.Meta["trend"] != "unknown" {
		t.Fatalf("expected unknown, got %s", res.Meta["trend"])
	}
}

func TestTrend_StableAllHealthy(t *testing.T) {
	tc := NewTrendChecker(&fixedChecker{StatusHealthy}, TrendConfig{Window: 6, MinSample: 3})
	for i := 0; i < 6; i++ {
		tc.Check(context.Background(), "svc")
	}
	res := tc.Check(context.Background(), "svc")
	if res.Meta["trend"] != "stable" {
		t.Fatalf("expected stable, got %s", res.Meta["trend"])
	}
}

func TestTrend_DegradingPattern(t *testing.T) {
	tc := NewTrendChecker(nil, TrendConfig{Window: 8, MinSample: 4})
	// manually populate history: first half healthy, second half unhealthy
	tc.inner = &fixedChecker{StatusUnhealthy}
	healthyInner := &fixedChecker{StatusHealthy}

	// inject 4 healthy results via a temporary inner
	tc.inner = healthyInner
	for i := 0; i < 4; i++ {
		tc.Check(context.Background(), "svc")
	}
	// inject 4 unhealthy results
	tc.inner = &fixedChecker{StatusUnhealthy}
	var res Result
	for i := 0; i < 4; i++ {
		res = tc.Check(context.Background(), "svc")
	}
	if res.Meta["trend"] != "degrading" {
		t.Fatalf("expected degrading, got %s", res.Meta["trend"])
	}
}

func TestTrend_DefaultConfigOnZero(t *testing.T) {
	tc := NewTrendChecker(&fixedChecker{StatusHealthy}, TrendConfig{})
	if tc.cfg.Window != DefaultTrendConfig().Window {
		t.Fatalf("expected default window %d, got %d", DefaultTrendConfig().Window, tc.cfg.Window)
	}
}

func TestTrendHandler_ReturnsJSON(t *testing.T) {
	tc := NewTrendChecker(&fixedChecker{StatusHealthy}, DefaultTrendConfig())
	h := NewTrendStatusHandler(tc)

	req := httptest.NewRequest(http.MethodGet, "/?service=svc", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var snap trendSnapshot
	if err := json.NewDecoder(rec.Body).Decode(&snap); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if snap.Service != "svc" {
		t.Fatalf("expected service svc, got %s", snap.Service)
	}
}

func TestTrendHandler_MissingService(t *testing.T) {
	tc := NewTrendChecker(&fixedChecker{StatusHealthy}, DefaultTrendConfig())
	h := NewTrendStatusHandler(tc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
