package health

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

// stubChecker is a simple Checker whose behaviour is controlled by fields.
type stubChecker struct {
	result Result
}

func (s *stubChecker) Check(_ context.Context, _ string) Result { return s.result }

func TestFallback_PrimarySucceeds(t *testing.T) {
	primary := &stubChecker{result: Result{Status: StatusHealthy}}
	fallback := &stubChecker{result: Result{Status: StatusUnhealthy}}

	fc := NewFallbackChecker(primary, fallback, slog.Default())
	res := fc.Check(context.Background(), "svc")

	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
}

func TestFallback_PrimaryUnhealthyNoFallback(t *testing.T) {
	primary := &stubChecker{result: Result{Status: StatusUnhealthy}}
	fallback := &stubChecker{result: Result{Status: StatusHealthy}}

	fc := NewFallbackChecker(primary, fallback, slog.Default())
	res := fc.Check(context.Background(), "svc")

	// Unhealthy (no error) must NOT trigger fallback.
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", res.Status)
	}
}

func TestFallback_PrimaryErrors_FallbackUsed(t *testing.T) {
	primary := &stubChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("dial error")}}
	fallback := &stubChecker{result: Result{Status: StatusHealthy}}

	fc := NewFallbackChecker(primary, fallback, slog.Default())
	res := fc.Check(context.Background(), "svc")

	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy from fallback, got %s", res.Status)
	}
	if res.Source != "fallback" {
		t.Fatalf("expected source=fallback, got %q", res.Source)
	}
}

func TestFallback_NilLogger(t *testing.T) {
	primary := &stubChecker{result: Result{Status: StatusHealthy}}
	fallback := &stubChecker{result: Result{Status: StatusHealthy}}

	// Should not panic with nil logger.
	fc := NewFallbackChecker(primary, fallback, nil)
	res := fc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("unexpected status %s", res.Status)
	}
}
