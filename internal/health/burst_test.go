package health

import (
	"context"
	"testing"
	"time"
)

func TestBurst_AllowsUnderLimit(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	c := NewBurstChecker(inner, BurstConfig{MaxBurst: 3, Window: time.Second})

	for i := 0; i < 3; i++ {
		r := c.Check(context.Background(), "svc")
		if r.Status != StatusHealthy {
			t.Fatalf("call %d: expected healthy, got %s", i, r.Status)
		}
	}
}

func TestBurst_RejectsOverLimit(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	c := NewBurstChecker(inner, BurstConfig{MaxBurst: 2, Window: time.Second})

	// First two should pass.
	for i := 0; i < 2; i++ {
		c.Check(context.Background(), "svc")
	}
	// Third should be rejected.
	r := c.Check(context.Background(), "svc")
	if r.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy on burst, got %s", r.Status)
	}
}

func TestBurst_DefaultConfigOnZero(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	c := NewBurstChecker(inner, BurstConfig{}) // zero values → defaults
	if c == nil {
		t.Fatal("expected non-nil checker")
	}
}

func TestBurst_WindowEviction(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	// Very short window so old entries expire quickly.
	c := NewBurstChecker(inner, BurstConfig{MaxBurst: 2, Window: 20 * time.Millisecond})

	// Fill the burst window.
	c.Check(context.Background(), "svc")
	c.Check(context.Background(), "svc")

	// Wait for the window to expire.
	time.Sleep(30 * time.Millisecond)

	// Now calls should be allowed again.
	r := c.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy after window eviction, got %s", r.Status)
	}
}

func TestBurst_NilLoggerUsesDefault(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	c := NewBurstChecker(inner, BurstConfig{MaxBurst: 5, Window: time.Second, Logger: nil})
	r := c.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", r.Status)
	}
}

func TestBurstHandler_ReturnsJSON(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	c := NewBurstChecker(inner, BurstConfig{MaxBurst: 4, Window: time.Second})

	// Perform one check so current_count == 1.
	c.Check(context.Background(), "svc")

	h := NewBurstStatusHandler(c)
	rr := serveHandler(h)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !contains(body, "max_burst") {
		t.Fatalf("expected max_burst in response: %s", body)
	}
}

func TestBurstHandler_NotABurstChecker(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	h := NewBurstStatusHandler(inner)
	rr := serveHandler(h)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
