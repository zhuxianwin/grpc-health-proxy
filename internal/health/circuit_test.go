package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func fastConfig() CircuitConfig {
	return CircuitConfig{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		OpenDuration:     50 * time.Millisecond,
	}
}

func TestCircuit_ClosedOnSuccess(t *testing.T) {
	stub := &stubChecker{result: Result{Status: StatusHealthy}}
	cb := NewCircuitBreaker(stub, fastConfig())
	res := cb.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %v", res.Status)
	}
}

func TestCircuit_OpensAfterThreshold(t *testing.T) {
	stub := &stubChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("down")}}
	cb := NewCircuitBreaker(stub, fastConfig())
	for i := 0; i < 3; i++ {
		cb.Check(context.Background(), "svc")
	}
	res := cb.Check(context.Background(), "svc")
	if res.Err == nil {
		t.Fatal("expected circuit open error")
	}
	// inner should not have been called on the 4th attempt
	if stub.calls > 3 {
		t.Fatalf("inner called %d times, want <=3", stub.calls)
	}
}

func TestCircuit_HalfOpenAfterDuration(t *testing.T) {
	stub := &stubChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("down")}}
	cfg := fastConfig()
	cb := NewCircuitBreaker(stub, cfg)
	for i := 0; i < 3; i++ {
		cb.Check(context.Background(), "svc")
	}
	time.Sleep(cfg.OpenDuration + 10*time.Millisecond)
	// Now stub returns healthy
	stub.result = Result{Status: StatusHealthy}
	res := cb.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy in half-open, got %v", res.Status)
	}
}

func TestCircuit_ClosesAfterSuccessThreshold(t *testing.T) {
	stub := &stubChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("down")}}
	cfg := fastConfig()
	cb := NewCircuitBreaker(stub, cfg)
	for i := 0; i < 3; i++ {
		cb.Check(context.Background(), "svc")
	}
	time.Sleep(cfg.OpenDuration + 10*time.Millisecond)
	stub.result = Result{Status: StatusHealthy}
	for i := 0; i < cfg.SuccessThreshold; i++ {
		cb.Check(context.Background(), "svc")
	}
	// Circuit should be closed; inner must be called
	prev := stub.calls
	cb.Check(context.Background(), "svc")
	if stub.calls == prev {
		t.Fatal("inner not called after circuit closed")
	}
}

func TestDefaultCircuitConfig(t *testing.T) {
	cfg := DefaultCircuitConfig()
	if cfg.FailureThreshold <= 0 || cfg.SuccessThreshold <= 0 || cfg.OpenDuration <= 0 {
		t.Fatalf("invalid default config: %+v", cfg)
	}
}

// stubChecker is a simple Checker used across circuit tests.
type stubChecker struct {
	result Result
	calls  int
}

func (s *stubChecker) Check(_ context.Context, _ string) Result {
	s.calls++
	return s.result
}
