package health

import (
	"context"
	"testing"
	"time"
)

func TestSpread_DelegatesResult(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewSpreadChecker(inner, SpreadConfig{MaxDelay: 10 * time.Millisecond}, nil)

	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %v", res.Status)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", inner.calls)
	}
}

func TestSpread_UnhealthyPropagated(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	c := NewSpreadChecker(inner, SpreadConfig{MaxDelay: 5 * time.Millisecond}, nil)

	res := c.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %v", res.Status)
	}
}

func TestSpread_ContextCancelledDuringDelay(t *testing.T) {
	// Use a very large delay so the context cancel fires first.
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewSpreadChecker(inner, SpreadConfig{MaxDelay: 10 * time.Second}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	res := c.Check(ctx, "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy on cancelled context, got %v", res.Status)
	}
	if res.Err == nil {
		t.Fatal("expected non-nil error on cancelled context")
	}
	if inner.calls != 0 {
		t.Fatalf("inner should not have been called, got %d calls", inner.calls)
	}
}

func TestSpread_DefaultConfigOnZero(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	// Zero MaxDelay should trigger default.
	c := NewSpreadChecker(inner, SpreadConfig{MaxDelay: 0}, nil)

	sc := c.(*spreadChecker)
	if sc.cfg.MaxDelay != DefaultSpreadConfig().MaxDelay {
		t.Fatalf("expected default MaxDelay %v, got %v", DefaultSpreadConfig().MaxDelay, sc.cfg.MaxDelay)
	}
}

func TestSpread_NilLoggerUsesDefault(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewSpreadChecker(inner, SpreadConfig{MaxDelay: 5 * time.Millisecond}, nil)
	sc := c.(*spreadChecker)
	if sc.logger == nil {
		t.Fatal("expected non-nil logger when nil passed")
	}
}
