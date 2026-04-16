package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWarmupHandler_ReturnsJSON(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	cfg := WarmupConfig{Duration: 5 * time.Second, MinChecks: 2}
	c := NewWarmupChecker(inner, cfg, nil)

	h := NewWarmupStatusHandler(c)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content-type: %s", ct)
	}

	var resp warmupStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.MinChecks != 2 {
		t.Fatalf("expected min_checks=2, got %d", resp.MinChecks)
	}
}

func TestWarmupHandler_WarmedUpAfterChecks(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	cfg := WarmupConfig{Duration: 0, MinChecks: 1}
	c := NewWarmupChecker(inner, cfg, nil)
	c.Check(nil, "svc") //nolint:staticcheck

	h := NewWarmupStatusHandler(c)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	var resp warmupStatusResponse
	json.NewDecoder(rec.Body).Decode(&resp) //nolint:errcheck
	if !resp.WarmedUp {
		t.Fatal("expected warmed_up=true")
	}
}

func TestWarmupHandler_NotAWarmupChecker(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	h := NewWarmupStatusHandler(inner)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rec.Code)
	}
}
