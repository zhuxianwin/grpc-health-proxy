package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"context"
	"time"
)

func TestQuotaWindowHandler_ReturnsJSON(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	cfg := QuotaWindowConfig{Window: time.Second, MaxCalls: 5}
	c := NewQuotaWindowChecker(inner, cfg)

	// Trigger a check so the service appears in the map.
	c.Check(context.Background(), "svc")

	h := NewQuotaWindowStatusHandler(c)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var statuses []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&statuses); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(statuses) == 0 {
		t.Fatal("expected at least one entry")
	}
}

func TestQuotaWindowHandler_ExceededTrue(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	cfg := QuotaWindowConfig{Window: time.Second, MaxCalls: 1}
	c := NewQuotaWindowChecker(inner, cfg)

	c.Check(context.Background(), "svc")
	c.Check(context.Background(), "svc") // exceeds

	h := NewQuotaWindowStatusHandler(c)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	var statuses []map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&statuses)
	if len(statuses) == 0 {
		t.Fatal("expected status entry")
	}
	exceeded, _ := statuses[0]["exceeded"].(bool)
	if !exceeded {
		t.Fatal("expected exceeded=true")
	}
}

func TestQuotaWindowHandler_NotAQuotaWindowChecker(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	h := NewQuotaWindowStatusHandler(inner) // not a quotaWindowChecker
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rec.Code)
	}
}
