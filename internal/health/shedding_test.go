package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func slowChecker(d time.Duration, status Status) Checker {
	return checkerFunc(func(_ context.Context, _ string) Result {
		time.Sleep(d)
		return Result{Status: status}
	})
}

func TestShedding_AllowsUnderLatency(t *testing.T) {
	cfg := SheddingConfig{
		MaxLatency:    200 * time.Millisecond,
		WindowSize:    5,
		SheddingRatio: 0.8,
	}
	// fast checker — well below threshold
	inner := slowChecker(0, StatusHealthy)
	c := NewSheddingChecker(inner, cfg, nil)

	for i := 0; i < 10; i++ {
		res := c.Check(context.Background(), "svc")
		if res.Status != StatusHealthy {
			t.Fatalf("expected healthy, got %v", res.Status)
		}
	}
}

func TestShedding_ShedsAfterHighLatency(t *testing.T) {
	cfg := SheddingConfig{
		MaxLatency:    10 * time.Millisecond,
		WindowSize:    4,
		SheddingRatio: 0.75, // 3 of 4 must exceed threshold
	}
	// slow checker — always above threshold
	inner := slowChecker(50*time.Millisecond, StatusHealthy)
	c := NewSheddingChecker(inner, cfg, nil)

	// fill the window with slow observations
	for i := 0; i < 4; i++ {
		c.Check(context.Background(), "svc") //nolint:errcheck
	}

	// next call should be shed
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy (shed), got %v", res.Status)
	}
	if !errors.Is(res.Err, ErrLoadShed) {
		t.Fatalf("expected ErrLoadShed, got %v", res.Err)
	}
}

func TestShedding_DefaultConfig(t *testing.T) {
	cfg := DefaultSheddingConfig()
	if cfg.WindowSize <= 0 {
		t.Fatal("WindowSize must be positive")
	}
	if cfg.SheddingRatio <= 0 || cfg.SheddingRatio > 1 {
		t.Fatal("SheddingRatio must be in (0,1]")
	}
	if cfg.MaxLatency <= 0 {
		t.Fatal("MaxLatency must be positive")
	}
}

func TestShedding_ZeroWindowUsesDefault(t *testing.T) {
	cfg := SheddingConfig{WindowSize: 0, SheddingRatio: 0, MaxLatency: 100 * time.Millisecond}
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		return Result{Status: StatusHealthy}
	})
	// Should not panic with zero config values.
	c := NewSheddingChecker(inner, cfg, nil)
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %v", res.Status)
	}
}
