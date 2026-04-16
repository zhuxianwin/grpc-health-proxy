package health

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// countingChecker records how many times Check is called.
type countingChecker struct {
	calls int64
	delay time.Duration
	result Result
}

func (c *countingChecker) Check(_ context.Context, _ string) Result {
	atomic.AddInt64(&c.calls, 1)
	if c.delay > 0 {
		time.Sleep(c.delay)
	}
	return c.result
}

func TestDedup_SingleCall(t *testing.T) {
	inner := &countingChecker{result: Healthy("svc")}
	d := NewDedupChecker(inner)

	r := d.Check(context.Background(), "svc")
	if !r.OK() {
		t.Fatalf("expected healthy, got %v", r)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls)
	}
}

func TestDedup_ConcurrentCallsDeduped(t *testing.T) {
	inner := &countingChecker{
		result: Healthy("svc"),
		delay:  40 * time.Millisecond,
	}
	d := NewDedupChecker(inner)

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			r := d.Check(context.Background(), "svc")
			if !r.OK() {
				t.Errorf("expected healthy result")
			}
		}()
	}
	wg.Wait()

	// All goroutines fired concurrently; upstream should have been hit far
	// fewer times than goroutines (ideally 1, but allow small variance).
	if inner.calls >= goroutines {
		t.Fatalf("dedup did not collapse calls: got %d upstream calls for %d goroutines",
			inner.calls, goroutines)
	}
}

func TestDedup_DifferentServicesNotMerged(t *testing.T) {
	inner := &countingChecker{result: Healthy("x")}
	d := NewDedupChecker(inner)

	d.Check(context.Background(), "alpha")
	d.Check(context.Background(), "beta")

	if inner.calls != 2 {
		t.Fatalf("expected 2 upstream calls for different services, got %d", inner.calls)
	}
}

func TestDedup_ForgetAllowsNewCall(t *testing.T) {
	inner := &countingChecker{result: Healthy("svc")}
	d := NewDedupChecker(inner)

	d.Check(context.Background(), "svc")
	d.Forget("svc")
	d.Check(context.Background(), "svc")

	if inner.calls != 2 {
		t.Fatalf("expected 2 calls after Forget, got %d", inner.calls)
	}
}

func TestDedup_NilPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner checker")
		}
	}()
	NewDedupChecker(nil)
}
