package health

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFallbackHandler_PrimaryHealthy(t *testing.T) {
	primary := &stubChecker{result: Result{Status: StatusHealthy}}
	fallback := &stubChecker{result: Result{Status: StatusUnhealthy}}
	fc := NewFallbackChecker(primary, fallback, slog.Default())

	h := NewFallbackStatusHandler(fc, "my-svc", slog.Default())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["fallback_used"].(bool) {
		t.Error("expected fallback_used=false")
	}
}

func TestFallbackHandler_FallbackUsed(t *testing.T) {
	primary := &stubChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("timeout")}}
	fallback := &stubChecker{result: Result{Status: StatusHealthy}}
	fc := NewFallbackChecker(primary, fallback, slog.Default())

	h := NewFallbackStatusHandler(fc, "my-svc", nil)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body["fallback_used"].(bool) {
		t.Error("expected fallback_used=true")
	}
	if body["source"].(string) != "fallback" {
		t.Errorf("expected source=fallback, got %s", body["source"])
	}
}

func TestFallbackHandler_Unhealthy(t *testing.T) {
	primary := &stubChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("down")}}
	fallback := &stubChecker{result: Result{Status: StatusUnhealthy}}
	fc := NewFallbackChecker(primary, fallback, slog.Default())

	h := NewFallbackStatusHandler(fc, "my-svc", slog.Default())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}
