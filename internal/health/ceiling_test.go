package health

import (
	"context"
	"errors"
	"testing"
)

func TestCeiling_BelowLimitDelegates(t *testing.T) {
	calls := 0
	inner := CheckerFunc(func(_ context.Context, _ string) Result {
		calls++
		return Result{Status: StatusUnhealthy}
	})
	c := NewCeilingChecker(inner, CeilingConfig{MaxUnhealthy: 3})

	// First three calls should all reach the inner checker.
	for i := 0; i < 3; i++ {
		r := c.Check(context.Background(), "svc")
		if r.Status != StatusUnhealthy {
			t.Fatalf("expected unhealthy, got %v", r.Status)
		}
	}
	if calls != 3 {
		t.Fatalf("expected 3 inner calls, got %d", calls)
	}
}

func TestCeiling_SuppressesAfterLimit(t *testing.T) {
	calls := 0
	inner := CheckerFunc(func(_ context.Context, _ string) Result {
		calls++
		return Result{Status: StatusUnhealthy}
	})
	c := NewCeilingChecker(inner, CeilingConfig{MaxUnhealthy: 2})

	// Exhaust ceiling.
	c.Check(context.Background(), "svc")
	c.Check(context.Background(), "svc")

	// Further calls must be suppressed.
	for i := 0; i < 5; i++ {
		r := c.Check(context.Background(), "svc")
		if r.Status != StatusUnhealthy {
			t.Fatalf("expected cached unhealthy, got %v", r.Status)
		}
	}
	if calls != 2 {
		t.Fatalf("expected exactly 2 inner calls, got %d", calls)
	}
}

func TestCeiling_ResetsOnHealthy(t *testing.T) {
	toggle := false
	inner := CheckerFunc(func(_ context.Context, _ string) Result {
		if toggle {
			return Result{Status: StatusHealthy}
		}
		return Result{Status: StatusUnhealthy}
	})
	c := NewCeilingChecker(inner, CeilingConfig{MaxUnhealthy: 1})

	c.Check(context.Background(), "svc") // unhealthy — ceiling hit
	toggle = true
	r := c.Check(context.Background(), "svc") // should bypass suppression since healthy resets
	// After reset the next call should delegate again.
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy after reset, got %v", r.Status)
	}
}

func TestCeiling_DefaultConfigOnZero(t *testing.T) {
	inner := CheckerFunc(func(_ context.Context, _ string) Result {
		return Result{Status: StatusUnhealthy, Err: errors.New("boom")}
	})
	// Zero MaxUnhealthy should fall back to default (1).
	c := NewCeilingChecker(inner, CeilingConfig{MaxUnhealthy: 0})
	c.Check(context.Background(), "svc")
	// Second call should be suppressed.
	calls := 0
	inner2 := CheckerFunc(func(_ context.Context, _ string) Result {
		calls++
		return Result{Status: StatusUnhealthy}
	})
	c2 := NewCeilingChecker(inner2, CeilingConfig{MaxUnhealthy: 0})
	c2.Check(context.Background(), "svc")
	c2.Check(context.Background(), "svc")
	if calls != 1 {
		t.Fatalf("expected 1 inner call after default ceiling of 1, got %d", calls)
	}
}

func TestCeiling_NilLoggerUsesDefault(t *testing.T) {
	inner := CheckerFunc(func(_ context.Context, _ string) Result {
		return Result{Status: StatusHealthy}
	})
	// Should not panic with nil logger.
	c := NewCeilingChecker(inner, CeilingConfig{Logger: nil, MaxUnhealthy: 2})
	r := c.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %v", r.Status)
	}
}
