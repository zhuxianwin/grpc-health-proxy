package health

import (
	"context"
	"testing"
	"time"
)

// fixedChecker always returns the same Result.
type fixedChecker struct{ result Result }

func (f *fixedChecker) Check(_ context.Context) Result { return f.result }

func TestRateLimitedChecker_AllowsUnderLimit(t *testing.T) {
	inner := &fixedChecker{result: Result{Status: StatusHealthy}}
	rl := NewRateLimitedChecker(inner, "svc", RateLimitConfig{MaxChecksPerSecond: 100, Burst: 10})

	for i := 0; i < 5; i++ {
		got := rl.Check(context.Background())
		if got.Status != StatusHealthy {
			t.Fatalf("call %d: expected healthy, got %v", i, got.Status)
		}
	}
}

func TestRateLimitedChecker_BlocksOverLimit(t *testing.T) {
	inner := &fixedChecker{result: Result{Status: StatusHealthy}}
	// Burst of 2, very low rate so tokens won't refill during the test.
	rl := NewRateLimitedChecker(inner, "svc", RateLimitConfig{MaxChecksPerSecond: 0.001, Burst: 2})

	// First two calls should succeed (burst).
	for i := 0; i < 2; i++ {
		if got := rl.Check(context.Background()); got.Status != StatusHealthy {
			t.Fatalf("expected healthy on call %d", i)
		}
	}

	// Third call must be rate-limited.
	got := rl.Check(context.Background())
	if got.Status != StatusUnknown {
		t.Fatalf("expected unknown (rate-limited), got %v", got.Status)
	}
	if got.Err == nil {
		t.Fatal("expected non-nil error when rate-limited")
	}
}

func TestRateLimitedChecker_TokensRefill(t *testing.T) {
	inner := &fixedChecker{result: Result{Status: StatusHealthy}}
	// 1000 rps so tokens refill quickly; burst of 1.
	rl := NewRateLimitedChecker(inner, "svc", RateLimitConfig{MaxChecksPerSecond: 1000, Burst: 1})

	// Exhaust the single burst token.
	rl.Check(context.Background())

	// Wait for at least one token to refill.
	time.Sleep(5 * time.Millisecond)

	got := rl.Check(context.Background())
	if got.Status != StatusHealthy {
		t.Fatalf("expected healthy after token refill, got %v (err=%v)", got.Status, got.Err)
	}
}

func TestDefaultRateLimitConfig(t *testing.T) {
	cfg := DefaultRateLimitConfig()
	if cfg.MaxChecksPerSecond <= 0 {
		t.Errorf("expected positive MaxChecksPerSecond, got %v", cfg.MaxChecksPerSecond)
	}
	if cfg.Burst <= 0 {
		t.Errorf("expected positive Burst, got %v", cfg.Burst)
	}
}
