package health

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSentinelHandler_EmptyWhenNoTrips(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewSentinelChecker(inner, DefaultSentinelConfig())

	h := NewSentinelStatusHandler(c)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out []sentinelStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty, got %v", out)
	}
}

func TestSentinelHandler_ReportsTrippedService(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("down")}}
	cfg := SentinelConfig{TripAfter: 1, ResetAfter: defaultResetAfter}
	c := NewSentinelChecker(inner, cfg)
	c.Check(testCtx(), "mysvc")

	h := NewSentinelStatusHandler(c)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	var out []sentinelStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("expected at least one entry")
	}
	if !out[0].Tripped {
		t.Errorf("expected tripped=true")
	}
}

func TestSentinelHandler_NotASentinelChecker(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	h := NewSentinelStatusHandler(inner)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
