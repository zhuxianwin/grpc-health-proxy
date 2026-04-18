package health

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type countingChecker struct {
	calls atomic.Int32
	result Result
	delay  time.Duration
}

func (c *countingChecker) Check(ctx context.Context, service string) Result {
	if c.delay > 0 {
		time.Sleep(c.delay)
	}
	c.calls.Add(1)
	return c.result
}

func TestCoalesce_SingleCall(t *testing.T) {
	inner := &countingChecker{result: Result{Status: StatusHealthy}}
	ch := NewCoalesceChecker(inner, DefaultCoalesceConfig(), nil)

	r := ch.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %v", r.Status)
	}
	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls.Load())
	}
}

func TestCoalesce_ConcurrentCallsCoalesced(t *testing.T) {
	inner := &countingChecker{
		result: Result{Status: StatusHealthy},
		delay:  20 * time.Millisecond,
	}
	cfg := CoalesceConfig{Window: 30 * time.Millisecond}
	ch := NewCoalesceChecker(inner, cfg, nil)

	const n = 8
	var wg sync.WaitGroup
	results := make([]Result, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = ch.Check(context.Background(), "svc")
		}(i)
	}
	wg.Wait()

	for i, r := range results {
		if r.Status != StatusHealthy {
			t.Errorf("goroutine %d: expected healthy", i)
		}
	}
	if inner.calls.Load() > 2 {
		t.Errorf("expected coalescing, got %d upstream calls", inner.calls.Load())
	}
}

func TestCoalesce_DifferentServicesNotMerged(t *testing.T) {
	inner := &countingChecker{result: Result{Status: StatusHealthy}}
	ch := NewCoalesceChecker(inner, DefaultCoalesceConfig(), nil)

	var wg sync.WaitGroup
	for _, svc := range []string{"a", "b", "c"} {
		wg.Add(1)
		s := svc
		go func() { defer wg.Done(); ch.Check(context.Background(), s) }()
	}
	wg.Wait()

	if inner.calls.Load() < 3 {
		t.Errorf("expected at least 3 calls for 3 services, got %d", inner.calls.Load())
	}
}

func TestCoalesce_ContextCancelledWhileWaiting(t *testing.T) {
	inner := &countingChecker{result: Result{Status: StatusHealthy}, delay: 200 * time.Millisecond}
	cfg := CoalesceConfig{Window: 100 * time.Millisecond}
	ch := NewCoalesceChecker(inner, cfg, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	r := ch.Check(ctx, "svc")
	if r.Err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestCoalesce_UnhealthyResultPropagated(t *testing.T) {
	inner := &countingChecker{result: Result{Status: StatusUnhealthy}}
	ch := NewCoalesceChecker(inner, DefaultCoalesceConfig(), nil)

	r := ch.Check(context.Background(), "svc")
	if r.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %v", r.Status)
	}
}
