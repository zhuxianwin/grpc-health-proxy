package health

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// countingChecker records how many times Check is called.
type countingChecker struct {
	calls   atomic.Int32
	results []Result // returned in order; last element repeated
}

func (c *countingChecker) Check(_ context.Context, _ string) Result {
	idx := int(c.calls.Add(1)) - 1
	if idx >= len(c.results) {
		idx = len(c.results) - 1
	}
	return c.results[idx]
}

func TestHedge_FirstSucceeds(t *testing.T) {
	inner := &countingChecker{results: []Result{{Status: StatusHealthy}}}
	h := NewHedgeChecker(inner, DefaultHedgeConfig(), nil)

	res := h.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls.Load())
	}
}

func TestHedge_HedgedRequestFires(t *testing.T) {
	errResult := Result{Status: StatusUnhealthy, Err: errors.New("timeout")}
	okResult := Result{Status: StatusHealthy}

	// first attempt errors, second (hedged) succeeds
	inner := &countingChecker{results: []Result{errResult, okResult}}
	cfg := HedgeConfig{Delay: 10 * time.Millisecond, MaxHedged: 2}
	h := NewHedgeChecker(inner, cfg, nil)

	res := h.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy from hedge, got %s", res.Status)
	}
}

func TestHedge_AllFail(t *testing.T) {
	sentinel := errors.New("all bad")
	inner := &countingChecker{
		results: []Result{{Status: StatusUnhealthy, Err: sentinel}},
	}
	cfg := HedgeConfig{Delay: 10 * time.Millisecond, MaxHedged: 2}
	h := NewHedgeChecker(inner, cfg, nil)

	res := h.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", res.Status)
	}
	if !errors.Is(res.Err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", res.Err)
	}
}

func TestHedge_ContextCancelled(t *testing.T) {
	blocking := &slowChecker{delay: 500 * time.Millisecond, result: Result{Status: StatusHealthy}}
	cfg := HedgeConfig{Delay: 10 * time.Millisecond, MaxHedged: 2}
	h := NewHedgeChecker(blocking, cfg, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	res := h.Check(ctx, "svc")
	if res.Err == nil {
		t.Fatal("expected error on context cancellation")
	}
}

func TestHedge_DefaultMaxHedgedZero(t *testing.T) {
	inner := &countingChecker{results: []Result{{Status: StatusHealthy}}}
	// MaxHedged = 0 should be corrected to default
	h := NewHedgeChecker(inner, HedgeConfig{Delay: 10 * time.Millisecond, MaxHedged: 0}, nil)
	res := h.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
}

// slowChecker sleeps for delay then returns result.
type slowChecker struct {
	delay  time.Duration
	result Result
}

func (s *slowChecker) Check(ctx context.Context, _ string) Result {
	select {
	case <-time.After(s.delay):
		return s.result
	case <-ctx.Done():
		return Result{Status: StatusUnhealthy, Err: ctx.Err()}
	}
}
