package health

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func testRLLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestRateLimitHandler_ReturnsJSON(t *testing.T) {
	inner := &fixedChecker{result: Result{Status: StatusHealthy}}
	rl := NewRateLimitedChecker(inner, "my-service", DefaultRateLimitConfig())

	handler := NewRateLimitStatusHandler(map[string]*RateLimitedChecker{
		"my-service": rl,
	}, testRLLogger())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ratelimit", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %q", ct)
	}

	var statuses []rateLimitServiceStatus
	if err := json.NewDecoder(rec.Body).Decode(&statuses); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status entry, got %d", len(statuses))
	}
	if statuses[0].Service != "my-service" {
		t.Errorf("unexpected service name %q", statuses[0].Service)
	}
}

func TestRateLimitHandler_TokenAvailableFalseWhenExhausted(t *testing.T) {
	inner := &fixedChecker{result: Result{Status: StatusHealthy}}
	rl := NewRateLimitedChecker(inner, "svc", RateLimitConfig{MaxChecksPerSecond: 0.001, Burst: 1})

	// Exhaust the single burst token.
	rl.Check(nil) //nolint:staticcheck

	handler := NewRateLimitStatusHandler(map[string]*RateLimitedChecker{"svc": rl}, testRLLogger())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ratelimit", nil)
	handler.ServeHTTP(rec, req)

	var statuses []rateLimitServiceStatus
	if err := json.NewDecoder(rec.Body).Decode(&statuses); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if statuses[0].TokenAvailable {
		t.Error("expected TokenAvailable=false after burst exhausted")
	}
}

func TestRateLimitHandler_NilLogger(t *testing.T) {
	handler := NewRateLimitStatusHandler(map[string]*RateLimitedChecker{}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ratelimit", nil)
	handler.ServeHTTP(rec, req) // must not panic
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
