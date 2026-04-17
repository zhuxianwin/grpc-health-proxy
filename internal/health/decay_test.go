package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDecay_PassesThroughWhenUnderThreshold(t *testing.T) {
	calls := 0
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		calls++
		return Result{Status: StatusUnhealthy}
	})
	cfg := DecayConfig{DecayAfter: 3, BaseDelay: 10 * time.Millisecond, MaxDelay: time.Second}
	c := NewDecayChecker(inner, cfg, nil)
	for i := 0; i < 2; i++ {
		c.Check(context.Background(), "svc")
	}
	if calls != 2 {
		t.Fatalf("expected 2 inner calls, got %d", calls)
	}
}

func TestDecay_CachesAfterThreshold(t *testing.T) {
	calls := 0
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		calls++
		return Result{Status: StatusUnhealthy, Err: errors.New("down")}
	})
	cfg := DecayConfig{DecayAfter: 2, BaseDelay: 500 * time.Millisecond, MaxDelay: 2 * time.Second}
	c := NewDecayChecker(inner, cfg, nil)
	// exhaust threshold
	c.Check(context.Background(), "svc")
	c.Check(context.Background(), "svc")
	before := calls
	// next call should be cached
	res := c.Check(context.Background(), "svc")
	if calls != before {
		t.Fatalf("expected cached result, inner called %d times", calls)
	}
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy cached result")
	}
}

func TestDecay_ResetsOnSuccess(t *testing.T) {
	healthy := false
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		if healthy {
			return Result{Status: StatusHealthy}
		}
		return Result{Status: StatusUnhealthy, Err: errors.New("down")}
	})
	cfg := DecayConfig{DecayAfter: 1, BaseDelay: 10 * time.Millisecond, MaxDelay: 50 * time.Millisecond}
	c := NewDecayChecker(inner, cfg, nil)
	c.Check(context.Background(), "svc") // fail #1 — triggers decay
	time.Sleep(15 * time.Millisecond)    // wait out cool-down
	healthy = true
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy after reset, got %s", res.Status)
	}
	// subsequent call must also go through (no decay)
	res2 := c.Check(context.Background(), "svc")
	if res2.Status != StatusHealthy {
		t.Fatalf("expected healthy on follow-up, got %s", res2.Status)
	}
}

func TestDecay_DefaultConfigOnZero(t *testing.T) {
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		return Result{Status: StatusHealthy}
	})
	c := NewDecayChecker(inner, DecayConfig{}, nil)
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
}
