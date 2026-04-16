package health

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type countingChecker struct {
	calls atomic.Int32
	result Result
}

func (c *countingChecker) Check(_ context.Context, _ string) Result {
	c.calls.Add(1)
	return c.result
}

func TestStale_FirstCallHitsInner(t *testing.T) {
	inner := &countingChecker{result: Healthy("svc")}
	sc := NewStaleChecker(inner, DefaultStaleConfig(), nil)

	r := sc.Check(context.Background(), "svc")
	if !r.Healthy {
		t.Fatal("expected healthy")
	}
	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls.Load())
	}
}

func TestStale_ServesCachedWithinWindow(t *testing.T) {
	inner := &countingChecker{result: Healthy("svc")}
	cfg := StaleConfig{StaleFor: 5 * time.Second}
	sc := NewStaleChecker(inner, cfg, nil)

	sc.Check(context.Background(), "svc") // prime cache
	time.Sleep(20 * time.Millisecond)      // let background goroutine settle

	_ = sc.Check(context.Background(), "svc") // should serve stale
	time.Sleep(30 * time.Millisecond)

	// inner called at most twice: initial + one revalidation
	if c := inner.calls.Load(); c > 2 {
		t.Fatalf("expected ≤2 inner calls, got %d", c)
	}
}

func TestStale_RevalidatesAfterExpiry(t *testing.T) {
	inner := &countingChecker{result: Healthy("svc")}
	cfg := StaleConfig{StaleFor: 30 * time.Millisecond}
	sc := NewStaleChecker(inner, cfg, nil)

	sc.Check(context.Background(), "svc")
	time.Sleep(60 * time.Millisecond) // let stale window expire

	sc.Check(context.Background(), "svc") // should call inner again
	time.Sleep(20 * time.Millisecond)

	if c := inner.calls.Load(); c < 2 {
		t.Fatalf("expected ≥2 inner calls after expiry, got %d", c)
	}
}

func TestStale_DefaultConfigOnZero(t *testing.T) {
	inner := &countingChecker{result: Healthy("svc")}
	sc := NewStaleChecker(inner, StaleConfig{}, nil)
	if sc.cfg.StaleFor != DefaultStaleConfig().StaleFor {
		t.Fatalf("expected default StaleFor, got %v", sc.cfg.StaleFor)
	}
}

func TestStale_DifferentServicesIndependent(t *testing.T) {
	inner := &countingChecker{result: Healthy("x")}
	sc := NewStaleChecker(inner, DefaultStaleConfig(), nil)

	sc.Check(context.Background(), "a")
	sc.Check(context.Background(), "b")

	if c := inner.calls.Load(); c != 2 {
		t.Fatalf("expected 2 calls for distinct services, got %d", c)
	}
}
