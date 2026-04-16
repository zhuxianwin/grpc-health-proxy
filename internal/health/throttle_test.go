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

func TestThrottle_PassesThroughFirstCall(t *testing.T) {
	inner := &countingChecker{result: Healthy("svc")}
	ch := NewThrottleChecker(inner, ThrottleConfig{MinInterval: 500 * time.Millisecond}, nil)

	res := ch.Check(context.Background(), "svc")
	if !res.OK() {
		t.Fatalf("expected healthy, got %v", res)
	}
	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 upstream call, got %d", inner.calls.Load())
	}
}

func TestThrottle_SuppressesSecondCallWithinInterval(t *testing.T) {
	inner := &countingChecker{result: Healthy("svc")}
	ch := NewThrottleChecker(inner, ThrottleConfig{MinInterval: 500 * time.Millisecond}, nil)

	ch.Check(context.Background(), "svc")
	ch.Check(context.Background(), "svc")

	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 upstream call, got %d", inner.calls.Load())
	}
}

func TestThrottle_AllowsCallAfterInterval(t *testing.T) {
	inner := &countingChecker{result: Healthy("svc")}
	ch := NewThrottleChecker(inner, ThrottleConfig{MinInterval: 20 * time.Millisecond}, nil)

	ch.Check(context.Background(), "svc")
	time.Sleep(30 * time.Millisecond)
	ch.Check(context.Background(), "svc")

	if inner.calls.Load() != 2 {
		t.Fatalf("expected 2 upstream calls, got %d", inner.calls.Load())
	}
}

func TestThrottle_DifferentServicesIndependent(t *testing.T) {
	inner := &countingChecker{result: Healthy("svc")}
	ch := NewThrottleChecker(inner, ThrottleConfig{MinInterval: 500 * time.Millisecond}, nil)

	ch.Check(context.Background(), "svc-a")
	ch.Check(context.Background(), "svc-b")

	if inner.calls.Load() != 2 {
		t.Fatalf("expected 2 upstream calls, got %d", inner.calls.Load())
	}
}

func TestThrottle_DefaultConfigUsedOnZeroInterval(t *testing.T) {
	inner := &countingChecker{result: Healthy("svc")}
	ch := NewThrottleChecker(inner, ThrottleConfig{MinInterval: 0}, nil)

	// Both calls within the default 2s window — second should be throttled.
	ch.Check(context.Background(), "svc")
	ch.Check(context.Background(), "svc")

	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 upstream call with default config, got %d", inner.calls.Load())
	}
}
