package health

import (
	"context"
	"testing"
)

type countingChecker struct {
	calls int
	result Result
}

func (c *countingChecker) Check(_ context.Context, _ string) Result {
	c.calls++
	return c.result
}

func TestSampling_AlwaysChecksAtRate1(t *testing.T) {
	inner := &countingChecker{result: Result{Status: StatusHealthy}}
	s := NewSamplingChecker(inner, SamplingConfig{Rate: 1.0}, nil)
	for i := 0; i < 10; i++ {
		s.Check(context.Background(), "svc")
	}
	if inner.calls != 10 {
		t.Fatalf("expected 10 calls, got %d", inner.calls)
	}
}

func TestSampling_NeverChecksAtRate0(t *testing.T) {
	inner := &countingChecker{result: Result{Status: StatusHealthy}}
	s := NewSamplingChecker(inner, SamplingConfig{Rate: 0.0}, nil)
	for i := 0; i < 20; i++ {
		s.Check(context.Background(), "svc")
	}
	if inner.calls != 0 {
		t.Fatalf("expected 0 calls, got %d", inner.calls)
	}
}

func TestSampling_ReturnsCachedOnSkip(t *testing.T) {
	unhealthy := Result{Status: StatusUnhealthy}
	inner := &countingChecker{result: unhealthy}
	// Force one real check first with rate 1, then switch to 0.
	s := NewSamplingChecker(inner, SamplingConfig{Rate: 1.0}, nil)
	s.Check(context.Background(), "svc")

	// Now create a rate-0 checker sharing the same last pointer via re-use.
	s2 := NewSamplingChecker(inner, SamplingConfig{Rate: 0.0}, nil)
	// Seed the cache manually.
	s2.last.Store(&unhealthy)
	r := s2.Check(context.Background(), "svc")
	if r.Status != StatusUnhealthy {
		t.Fatalf("expected cached unhealthy, got %v", r.Status)
	}
}

func TestSampling_DefaultConfig(t *testing.T) {
	cfg := DefaultSamplingConfig()
	if cfg.Rate != 1.0 {
		t.Fatalf("expected default rate 1.0, got %v", cfg.Rate)
	}
}

func TestSampling_ClampsRateAbove1(t *testing.T) {
	inner := &countingChecker{result: Result{Status: StatusHealthy}}
	s := NewSamplingChecker(inner, SamplingConfig{Rate: 5.0}, nil)
	if s.rate != 1.0 {
		t.Fatalf("expected clamped rate 1.0, got %v", s.rate)
	}
}
