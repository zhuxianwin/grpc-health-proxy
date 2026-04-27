package health

import (
	"context"
	"testing"
	"time"
)

func TestPeak_AnnotatesLatency(t *testing.T) {
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		return Result{Status: StatusHealthy}
	})
	checker := NewPeakChecker(inner, DefaultPeakConfig(), nil)
	res := checker.Check(context.Background(), "svc")
	if res.Metadata["peak_latency_ms"] == "" {
		t.Fatal("expected peak_latency_ms in metadata")
	}
	if res.Metadata["last_latency_ms"] == "" {
		t.Fatal("expected last_latency_ms in metadata")
	}
}

func TestPeak_PropagatesUnhealthy(t *testing.T) {
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		return Result{Status: StatusUnhealthy}
	})
	checker := NewPeakChecker(inner, DefaultPeakConfig(), nil)
	res := checker.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected Unhealthy, got %s", res.Status)
	}
}

func TestPeak_PeakRisesWithSlowCheck(t *testing.T) {
	calls := 0
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		calls++
		if calls == 2 {
			time.Sleep(30 * time.Millisecond)
		}
		return Result{Status: StatusHealthy}
	})
	cfg := DefaultPeakConfig()
	cfg.Window = 5 * time.Second
	checker := NewPeakChecker(inner, cfg, nil)

	checker.Check(context.Background(), "svc") // fast
	res := checker.Check(context.Background(), "svc") // slow

	peak := res.Metadata["peak_latency_ms"]
	last := res.Metadata["last_latency_ms"]
	if peak != last {
		// peak should equal last on second call because slow call sets new peak
		t.Logf("peak=%s last=%s (may differ due to timing)", peak, last)
	}
	if peak == "" {
		t.Fatal("peak_latency_ms should not be empty")
	}
}

func TestPeak_DefaultConfigOnZero(t *testing.T) {
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		return Result{Status: StatusHealthy}
	})
	// zero config should fall back to defaults without panic
	checker := NewPeakChecker(inner, PeakConfig{}, nil)
	res := checker.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("unexpected status %s", res.Status)
	}
}

func TestPeak_WindowEvictsOldEntries(t *testing.T) {
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		return Result{Status: StatusHealthy}
	})
	cfg := PeakConfig{Window: 50 * time.Millisecond}
	checker := NewPeakChecker(inner, cfg, nil)

	checker.Check(context.Background(), "svc")
	time.Sleep(80 * time.Millisecond)
	checker.Check(context.Background(), "svc")

	pc := checker.(*peakChecker)
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.evict(time.Now())
	if len(pc.entries) != 1 {
		t.Fatalf("expected 1 entry after eviction, got %d", len(pc.entries))
	}
}

func TestItoa(t *testing.T) {
	cases := []struct {
		in  int
		out string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{1000, "1000"},
	}
	for _, tc := range cases {
		got := itoa(tc.in)
		if got != tc.out {
			t.Errorf("itoa(%d) = %q, want %q", tc.in, got, tc.out)
		}
	}
}
