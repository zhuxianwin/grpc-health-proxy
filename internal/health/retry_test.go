package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

// stubChecker is a Checker that returns pre-canned results in order.
type stubChecker struct {
	results []Result
	calls   int
}

func (s *stubChecker) Check(_ context.Context, _ string) Result {
	idx := s.calls
	if idx >= len(s.results) {
		idx = len(s.results) - 1
	}
	s.calls++
	return s.results[idx]
}

func TestRetryChecker_SucceedsFirstAttempt(t *testing.T) {
	stub := &stubChecker{results: []Result{{Status: StatusHealthy}}}
	checker := NewRetryChecker(stub, DefaultRetryConfig())

	result := checker.Check(context.Background(), "svc")

	if result.Status != StatusHealthy {
		t.Fatalf("expected Healthy, got %s", result.Status)
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 call, got %d", stub.calls)
	}
}

func TestRetryChecker_RetriesOnError(t *testing.T) {
	errTransient := errors.New("connection refused")
	stub := &stubChecker{
		results: []Result{
			{Status: StatusUnknown, Err: errTransient},
			{Status: StatusUnknown, Err: errTransient},
			{Status: StatusHealthy},
		},
	}
	cfg := RetryConfig{Attempts: 3, Delay: time.Millisecond}
	checker := NewRetryChecker(stub, cfg)

	result := checker.Check(context.Background(), "svc")

	if result.Status != StatusHealthy {
		t.Fatalf("expected Healthy after retries, got %s", result.Status)
	}
	if stub.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", stub.calls)
	}
}

func TestRetryChecker_ExhaustsAttempts(t *testing.T) {
	errTransient := errors.New("timeout")
	stub := &stubChecker{
		results: []Result{
			{Status: StatusUnknown, Err: errTransient},
		},
	}
	cfg := RetryConfig{Attempts: 3, Delay: time.Millisecond}
	checker := NewRetryChecker(stub, cfg)

	result := checker.Check(context.Background(), "svc")

	if result.Err == nil {
		t.Fatal("expected error after exhausting attempts")
	}
	if stub.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", stub.calls)
	}
}

func TestRetryChecker_DoesNotRetryUnhealthy(t *testing.T) {
	stub := &stubChecker{
		results: []Result{
			{Status: StatusUnhealthy, Err: nil},
		},
	}
	cfg := RetryConfig{Attempts: 3, Delay: time.Millisecond}
	checker := NewRetryChecker(stub, cfg)

	result := checker.Check(context.Background(), "svc")

	if result.Status != StatusUnhealthy {
		t.Fatalf("expected Unhealthy, got %s", result.Status)
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 call (no retry for explicit unhealthy), got %d", stub.calls)
	}
}

func TestRetryChecker_RespectsContextCancellation(t *testing.T) {
	errTransient := errors.New("timeout")
	stub := &stubChecker{
		results: []Result{
			{Status: StatusUnknown, Err: errTransient},
		},
	}
	cfg := RetryConfig{Attempts: 5, Delay: 500 * time.Millisecond}
	checker := NewRetryChecker(stub, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result := checker.Check(ctx, "svc")

	if result.Err == nil {
		t.Fatal("expected context error")
	}
}
