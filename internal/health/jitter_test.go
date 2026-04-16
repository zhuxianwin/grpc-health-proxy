package health

import (
	"context"
	"testing"
	"time"
)

func TestJitter_DelegatesResult(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewJitterChecker(inner, JitterConfig{MaxJitter: 5 * time.Millisecond}, nil)
	r := c.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", r.Status)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", inner.calls)
	}
}

func TestJitter_RespectsContextCancel(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewJitterChecker(inner, JitterConfig{MaxJitter: 10 * time.Second}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := c.Check(ctx, "svc")
	if r.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy on cancelled ctx, got %s", r.Status)
	}
	if r.Err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestJitter_DefaultConfigOnZero(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewJitterChecker(inner, JitterConfig{MaxJitter: 0}, nil)
	j := c.(*jitterChecker)
	if j.cfg.MaxJitter != DefaultJitterConfig().MaxJitter {
		t.Fatalf("expected default jitter, got %v", j.cfg.MaxJitter)
	}
}

func TestJitter_UnhealthyPropagated(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	c := NewJitterChecker(inner, JitterConfig{MaxJitter: 2 * time.Millisecond}, nil)
	r := c.Check(context.Background(), "svc")
	if r.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", r.Status)
	}
}

// stubChecker is a minimal Checker used in jitter tests.
type stubChecker struct {
	result Result
	calls  int
}

func (s *stubChecker) Check(_ context.Context, _ string) Result {
	s.calls++
	return s.result
}
