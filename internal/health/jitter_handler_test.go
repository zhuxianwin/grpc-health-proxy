package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestJitterHandler_ReturnsJSON(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewJitterChecker(inner, JitterConfig{MaxJitter: 30 * time.Millisecond}, nil)
	h := NewJitterStatusHandler(c)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["max_jitter"] == "" {
		t.Fatal("expected max_jitter in response")
	}
}

func TestJitterHandler_ContentType(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewJitterChecker(inner, JitterConfig{MaxJitter: 10 * time.Millisecond}, nil)
	h := NewJitterStatusHandler(c)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestJitterHandler_NotAJitterChecker(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	h := NewJitterStatusHandler(inner)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
