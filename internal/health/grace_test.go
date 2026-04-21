package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGrace_SuppressesUnhealthyDuringWindow(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy, Message: "down"}}
	cfg := GraceConfig{Window: 5 * time.Second, MinChecks: 10}
	g := NewGraceChecker(inner, cfg)

	result := g.Check(context.Background(), "svc")
	if result.Status != StatusHealthy {
		t.Fatalf("expected healthy during grace, got %s", result.Status)
	}
}

func TestGrace_PassesThroughAfterWindow(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy, Message: "down"}}
	cfg := GraceConfig{Window: time.Millisecond, MinChecks: 1}
	g := NewGraceChecker(inner, cfg)

	time.Sleep(5 * time.Millisecond)

	// Run enough checks to satisfy MinChecks.
	var result Result
	for i := 0; i < 2; i++ {
		result = g.Check(context.Background(), "svc")
	}
	if result.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy after grace, got %s", result.Status)
	}
}

func TestGrace_HealthyPassedThrough(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	cfg := GraceConfig{Window: 10 * time.Second, MinChecks: 5}
	g := NewGraceChecker(inner, cfg)

	result := g.Check(context.Background(), "svc")
	if result.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", result.Status)
	}
}

func TestGrace_DefaultConfigOnZero(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	g := NewGraceChecker(inner, GraceConfig{})
	gc := g.(*graceChecker)
	if gc.cfg.Window != DefaultGraceConfig().Window {
		t.Fatalf("expected default window, got %s", gc.cfg.Window)
	}
	if gc.cfg.MinChecks != DefaultGraceConfig().MinChecks {
		t.Fatalf("expected default min checks, got %d", gc.cfg.MinChecks)
	}
}

func TestGraceHandler_ReturnsJSON(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	g := NewGraceChecker(inner, GraceConfig{Window: 2 * time.Second, MinChecks: 2})

	_ = g.Check(context.Background(), "svc")

	h := NewGraceStatusHandler(g)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp graceStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.InGrace {
		t.Error("expected in_grace to be true")
	}
	if resp.MinChecks != 2 {
		t.Errorf("expected min_checks 2, got %d", resp.MinChecks)
	}
}

func TestGraceHandler_NotAGraceChecker(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	h := NewGraceStatusHandler(inner)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
