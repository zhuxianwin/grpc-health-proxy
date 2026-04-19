package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFilter_AllowsAllWhenEmpty(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	cfg := DefaultFilterConfig()
	fc := NewFilterChecker(inner, cfg)

	r := fc.Check(context.Background(), "any.Service")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", r.Status)
	}
}

func TestFilter_BlocksUnlistedService(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	cfg := DefaultFilterConfig()
	cfg.AllowedServices = map[string]struct{}{"allowed.Service": {}}
	fc := NewFilterChecker(inner, cfg)

	r := fc.Check(context.Background(), "blocked.Service")
	if r.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy for blocked service, got %s", r.Status)
	}
}

func TestFilter_AllowsListedService(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	cfg := DefaultFilterConfig()
	cfg.AllowedServices = map[string]struct{}{"allowed.Service": {}}
	fc := NewFilterChecker(inner, cfg)

	r := fc.Check(context.Background(), "allowed.Service")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy for allowed service, got %s", r.Status)
	}
}

func TestFilter_NilLoggerUsesDefault(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	cfg := DefaultFilterConfig()
	cfg.Logger = nil
	fc := NewFilterChecker(inner, cfg)
	if fc.logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestFilterHandler_ReturnsJSON(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	cfg := DefaultFilterConfig()
	cfg.AllowedServices = map[string]struct{}{"svc.A": {}}
	fc := NewFilterChecker(inner, cfg)

	h := NewFilterStatusHandler(fc)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body struct {
		AllowAll bool     `json:"allow_all"`
		Services []string `json:"allowed_services"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body.AllowAll {
		t.Fatal("expected allow_all=false")
	}
	if len(body.Services) != 1 || body.Services[0] != "svc.A" {
		t.Fatalf("unexpected services: %v", body.Services)
	}
}

func TestFilterHandler_NotAFilterChecker(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	h := NewFilterStatusHandler(inner)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
