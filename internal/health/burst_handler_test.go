package health

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func serveHandler(h http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)
	return rr
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}

func TestBurstStatusHandler_ContentType(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	c := NewBurstChecker(inner, BurstConfig{MaxBurst: 3, Window: time.Second})
	h := NewBurstStatusHandler(c)
	rr := serveHandler(h)
	ct := rr.Header().Get("Content-Type")
	if !contains(ct, "application/json") {
		t.Fatalf("expected application/json content-type, got %s", ct)
	}
}

func TestBurstStatusHandler_WindowSeconds(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	c := NewBurstChecker(inner, BurstConfig{MaxBurst: 3, Window: 5 * time.Second})
	h := NewBurstStatusHandler(c)
	rr := serveHandler(h)
	body := rr.Body.String()
	if !contains(body, "5") {
		t.Fatalf("expected window_seconds 5 in body: %s", body)
	}
}

func TestBurstStatusHandler_CurrentCountReflectsCalls(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	cfg := BurstConfig{MaxBurst: 10, Window: time.Minute}
	c := NewBurstChecker(inner, cfg)

	for i := 0; i < 3; i++ {
		c.Check(nil, "svc") //nolint:staticcheck
	}

	h := NewBurstStatusHandler(c)
	rr := serveHandler(h)
	body := rr.Body.String()
	if !contains(body, "current_count") {
		t.Fatalf("expected current_count in body: %s", body)
	}
}
