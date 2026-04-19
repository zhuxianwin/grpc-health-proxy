package health

import (
	"context"
	"testing"
	"time"
)

func TestSticky_HealthyPassesThrough(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewStickyChecker(inner, StickyConfig{UnhealthyTTL: time.Second})
	res, err := c.Check(context.Background(), "svc")
	if err != nil || res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %v %v", res.Status, err)
	}
}

func TestSticky_UnhealthyIsHeld(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	c := NewStickyChecker(inner, StickyConfig{UnhealthyTTL: time.Hour})
	// first call records sticky
	c.Check(context.Background(), "svc") //nolint
	// make inner healthy now
	inner.result = Result{Status: StatusHealthy}
	// should still return unhealthy from sticky
	res, _ := c.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected sticky unhealthy, got %v", res.Status)
	}
}

func TestSticky_ClearsAfterTTL(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	c := NewStickyChecker(inner, StickyConfig{UnhealthyTTL: time.Millisecond})
	c.Check(context.Background(), "svc") //nolint
	time.Sleep(5 * time.Millisecond)
	inner.result = Result{Status: StatusHealthy}
	res, _ := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy after TTL, got %v", res.Status)
	}
}

func TestSticky_DefaultConfigOnZero(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewStickyChecker(inner, StickyConfig{})
	sc := c.(*stickyChecker)
	if sc.cfg.UnhealthyTTL != DefaultStickyConfig().UnhealthyTTL {
		t.Fatalf("expected default TTL")
	}
}

func TestSticky_HealthyAfterUnhealthyClearsEntry(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	c := NewStickyChecker(inner, StickyConfig{UnhealthyTTL: time.Millisecond})
	c.Check(context.Background(), "svc") //nolint
	time.Sleep(5 * time.Millisecond)
	inner.result = Result{Status: StatusHealthy}
	c.Check(context.Background(), "svc") //nolint
	sc := c.(*stickyChecker)
	sc.mu.Lock()
	_, exists := sc.entries["svc"]
	sc.mu.Unlock()
	if exists {
		t.Fatal("expected entry to be cleared after healthy result")
	}
}

func TestStickyHandler_ReturnsJSON(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	c := NewStickyChecker(inner, StickyConfig{UnhealthyTTL: time.Hour})
	c.Check(context.Background(), "svc") //nolint
	h := NewStickyStatusHandler(c)
	rec, req := newRecorder(), newRequest()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}

func TestStickyHandler_NotAStickyChecker(t *testing.T) {
	h := NewStickyStatusHandler(&stubChecker{})
	rec, req := newRecorder(), newRequest()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
