package health

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}

func TestCircuitHandler_Closed(t *testing.T) {
	stub := &stubChecker{result: Result{Status: StatusHealthy}}
	cb := NewCircuitBreaker(stub, fastConfig())
	h := NewCircuitStatusHandler("my-svc", cb, testLogger())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp CircuitStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.State != "closed" {
		t.Errorf("expected closed, got %q", resp.State)
	}
	if resp.Service != "my-svc" {
		t.Errorf("unexpected service %q", resp.Service)
	}
}

func TestCircuitHandler_Open(t *testing.T) {
	stub := &stubChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("down")}}
	cfg := fastConfig()
	cb := NewCircuitBreaker(stub, cfg)
	ctx := t.Context()
	for i := 0; i < cfg.FailureThreshold; i++ {
		cb.Check(ctx, "svc")
	}
	h := NewCircuitStatusHandler("svc", cb, testLogger())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	var resp CircuitStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.State != "open" {
		t.Errorf("expected open, got %q", resp.State)
	}
}

func TestCircuitHandler_NotABreaker(t *testing.T) {
	stub := &stubChecker{result: Result{Status: StatusHealthy}}
	h := NewCircuitStatusHandler("svc", stub, testLogger())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
