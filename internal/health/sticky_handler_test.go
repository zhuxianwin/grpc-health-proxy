package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStickyHandler_StuckFieldTrue(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	c := NewStickyChecker(inner, StickyConfig{UnhealthyTTL: time.Hour})
	c.Check(context.Background(), "alpha") //nolint

	h := NewStickyStatusHandler(c)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `"stuck":true`) {
		t.Fatalf("expected stuck:true in body, got: %s", body)
	}
	if !strings.Contains(body, `"stuck_until"`) {
		t.Fatalf("expected stuck_until in body, got: %s", body)
	}
}

func TestStickyHandler_EmptyWhenNoEntries(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewStickyChecker(inner, StickyConfig{UnhealthyTTL: time.Hour})

	h := NewStickyStatusHandler(c)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	body := strings.TrimSpace(rec.Body.String())
	if body != "null" && body != "[]" {
		t.Fatalf("expected empty list, got: %s", body)
	}
}

func TestStickyHandler_StuckFalseAfterTTL(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	c := NewStickyChecker(inner, StickyConfig{UnhealthyTTL: time.Millisecond})
	c.Check(context.Background(), "beta") //nolint
	time.Sleep(10 * time.Millisecond)

	h := NewStickyStatusHandler(c)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, `"stuck":true`) {
		t.Fatalf("expected stuck:false after TTL, got: %s", body)
	}
}
