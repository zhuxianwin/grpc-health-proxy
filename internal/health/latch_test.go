package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLatch_PassesThroughWhileHealthy(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	l := NewLatchChecker(inner, DefaultLatchConfig(), nil)
	res := l.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
}

func TestLatch_TripsAfterThreshold(t *testing.T) {
	cfg := LatchConfig{UnhealthyThreshold: 2, ResetAfter: 0}
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	l := NewLatchChecker(inner, cfg, nil).(*latchChecker)

	l.Check(context.Background(), "svc")
	if l.Tripped() {
		t.Fatal("should not be tripped after 1 failure")
	}
	l.Check(context.Background(), "svc")
	if !l.Tripped() {
		t.Fatal("should be tripped after 2 failures")
	}
}

func TestLatch_TrippedReturnsUnhealthyWithoutDelegating(t *testing.T) {
	cfg := LatchConfig{UnhealthyThreshold: 1, ResetAfter: 0}
	calls := 0
	inner := checkerFunc(func(_ context.Context, _ string) Result {
		calls++
		return Result{Status: StatusUnhealthy}
	})
	l := NewLatchChecker(inner, cfg, nil)
	l.Check(context.Background(), "svc") // trips
	l.Check(context.Background(), "svc") // should NOT delegate
	if calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", calls)
	}
}

func TestLatch_ResetRestoresDelegation(t *testing.T) {
	cfg := LatchConfig{UnhealthyThreshold: 1, ResetAfter: 0}
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	l := NewLatchChecker(inner, cfg, nil).(*latchChecker)
	l.Check(context.Background(), "svc")
	l.Reset()
	if l.Tripped() {
		t.Fatal("latch should be clear after Reset")
	}
}

func TestLatch_AutoResetAfterDuration(t *testing.T) {
	cfg := LatchConfig{UnhealthyThreshold: 1, ResetAfter: 20 * time.Millisecond}
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	l := NewLatchChecker(inner, cfg, nil).(*latchChecker)
	l.Check(context.Background(), "svc")
	if !l.Tripped() {
		t.Fatal("should be tripped")
	}
	time.Sleep(30 * time.Millisecond)
	// next Check should auto-reset and delegate
	l.Check(context.Background(), "svc")
}

func TestLatch_ErrLatchTripped(t *testing.T) {
	cfg := LatchConfig{UnhealthyThreshold: 1, ResetAfter: 0}
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	l := NewLatchChecker(inner, cfg, nil)
	l.Check(context.Background(), "svc")
	res := l.Check(context.Background(), "svc")
	if !errors.Is(res.Err, ErrLatchTripped) {
		t.Fatalf("expected ErrLatchTripped, got %v", res.Err)
	}
}

func TestLatch_DefaultConfigOnZero(t *testing.T) {
	cfg := LatchConfig{}
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	l := NewLatchChecker(inner, cfg, nil)
	if l == nil {
		t.Fatal("expected non-nil checker")
	}
}
