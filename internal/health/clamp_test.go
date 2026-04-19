package health

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClamp_NoAdjustmentNeeded(t *testing.T) {
	inner := staticChecker(Result{Status: StatusHealthy})
	c := NewClampChecker(inner, ClampConfig{MinStatus: StatusHealthy, MaxStatus: StatusUnhealthy})
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
}

func TestClamp_RaisesFloor(t *testing.T) {
	// inner returns unhealthy; clamp floor is healthy → should be raised
	inner := staticChecker(Result{Status: StatusUnhealthy})
	c := NewClampChecker(inner, ClampConfig{MinStatus: StatusHealthy, MaxStatus: StatusUnhealthy})
	res := c.Check(context.Background(), "svc")
	// StatusHealthy < StatusUnhealthy numerically, so floor raises unhealthy → healthy
	if res.Status != StatusHealthy {
		t.Fatalf("expected floor clamp to healthy, got %s", res.Status)
	}
}

func TestClamp_LowersCeiling(t *testing.T) {
	// Force a ceiling below healthy by setting max = StatusUnknown (0)
	inner := staticChecker(Result{Status: StatusHealthy})
	c := NewClampChecker(inner, ClampConfig{MinStatus: StatusUnknown, MaxStatus: StatusUnknown})
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusUnknown {
		t.Fatalf("expected ceiling clamp to unknown, got %s", res.Status)
	}
}

func TestClamp_DefaultConfigOnZero(t *testing.T) {
	inner := staticChecker(Result{Status: StatusHealthy})
	c := NewClampChecker(inner, ClampConfig{})
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("unexpected status %s", res.Status)
	}
}

func TestClampHandler_ReturnsJSON(t *testing.T) {
	inner := staticChecker(Result{Status: StatusHealthy})
	c := NewClampChecker(inner, ClampConfig{MinStatus: StatusHealthy, MaxStatus: StatusUnhealthy})
	h := NewClampStatusHandler(c)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "min_status") {
		t.Fatalf("expected min_status in body: %s", body)
	}
}

func TestClampHandler_NotAClampChecker(t *testing.T) {
	inner := staticChecker(Result{Status: StatusHealthy})
	h := NewClampStatusHandler(inner)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
