package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWarmup_SuppressesUnhealthyDuringWindow(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy, Message: "down"}}
	cfg := WarmupConfig{Duration: 1 * time.Hour, MinChecks: 1}
	c := NewWarmupChecker(inner, cfg, nil)

	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy during warmup, got %s", res.Status)
	}
	if res.Message != "warming up" {
		t.Fatalf("unexpected message: %s", res.Message)
	}
}

func TestWarmup_PassesThroughAfterWindow(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy, Message: "down"}}
	cfg := WarmupConfig{Duration: 0, MinChecks: 0}
	c := NewWarmupChecker(inner, cfg, nil)

	res := c.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy after warmup, got %s", res.Status)
	}
}

func TestWarmup_RequiresMinChecks(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	cfg := WarmupConfig{Duration: 0, MinChecks: 3}
	c := NewWarmupChecker(inner, cfg, nil)

	// Only 2 successful checks — still in warmup for unhealthy.
	c.Check(context.Background(), "svc")
	c.Check(context.Background(), "svc")

	// Now simulate failure — should be suppressed.
	wc := c.(*warmupChecker)
	wc.inner = &stubChecker{result: Result{Status: StatusUnhealthy}}
	res := wc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected suppressed, got %s", res.Status)
	}
}

func TestWarmup_HealthyPassedThrough(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy, Message: "ok"}}
	cfg := DefaultWarmupConfig()
	c := NewWarmupChecker(inner, cfg, nil)

	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy || res.Message != "ok" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestWarmup_ErrorSuppressedDuringWindow(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("conn refused")}}
	cfg := WarmupConfig{Duration: 1 * time.Hour, MinChecks: 1}
	c := NewWarmupChecker(inner, cfg, nil)

	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy during warmup, got %s", res.Status)
	}
}

// stubChecker is a simple test double.
type stubChecker struct {
	result Result
}

func (s *stubChecker) Check(_ context.Context, _ string) Result { return s.result }
