package health

import (
	"context"
	"testing"
	"time"
)

func TestQuotaWindow_AllowsUnderLimit(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	cfg := QuotaWindowConfig{Window: time.Second, MaxCalls: 3}
	c := NewQuotaWindowChecker(inner, cfg)

	for i := 0; i < 3; i++ {
		r := c.Check(context.Background(), "svc")
		if r.Status != StatusHealthy {
			t.Fatalf("call %d: expected healthy, got %s", i, r.Status)
		}
	}
	if inner.calls != 3 {
		t.Fatalf("expected 3 inner calls, got %d", inner.calls)
	}
}

func TestQuotaWindow_RejectsOverLimit(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	cfg := QuotaWindowConfig{Window: time.Second, MaxCalls: 2}
	c := NewQuotaWindowChecker(inner, cfg)

	c.Check(context.Background(), "svc")
	c.Check(context.Background(), "svc")
	r := c.Check(context.Background(), "svc") // third call — over limit
	if r.Status != StatusUnknown {
		t.Fatalf("expected unknown, got %s", r.Status)
	}
	if r.Err != ErrQuotaExceeded {
		t.Fatalf("expected ErrQuotaExceeded, got %v", r.Err)
	}
	if inner.calls != 2 {
		t.Fatalf("expected 2 inner calls, got %d", inner.calls)
	}
}

func TestQuotaWindow_DefaultConfigOnZero(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	c := NewQuotaWindowChecker(inner, QuotaWindowConfig{})
	qw := c.(*quotaWindowChecker)
	def := DefaultQuotaWindowConfig()
	if qw.cfg.Window != def.Window {
		t.Fatalf("expected default window %v, got %v", def.Window, qw.cfg.Window)
	}
	if qw.cfg.MaxCalls != def.MaxCalls {
		t.Fatalf("expected default max_calls %d, got %d", def.MaxCalls, qw.cfg.MaxCalls)
	}
}

func TestQuotaWindow_WindowEviction(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	// Very short window so entries expire quickly.
	cfg := QuotaWindowConfig{Window: 20 * time.Millisecond, MaxCalls: 2}
	c := NewQuotaWindowChecker(inner, cfg)

	c.Check(context.Background(), "svc")
	c.Check(context.Background(), "svc")

	// Wait for the window to expire.
	time.Sleep(30 * time.Millisecond)

	// Should be allowed again after eviction.
	r := c.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy after window expiry, got %s", r.Status)
	}
}

func TestQuotaWindow_DifferentServicesIndependent(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	cfg := QuotaWindowConfig{Window: time.Second, MaxCalls: 1}
	c := NewQuotaWindowChecker(inner, cfg)

	r1 := c.Check(context.Background(), "svc-a")
	r2 := c.Check(context.Background(), "svc-b")
	if r1.Status != StatusHealthy {
		t.Fatalf("svc-a: expected healthy, got %s", r1.Status)
	}
	if r2.Status != StatusHealthy {
		t.Fatalf("svc-b: expected healthy, got %s", r2.Status)
	}
}
