package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSentinel_DelegatesWhenHealthy(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewSentinelChecker(inner, DefaultSentinelConfig())
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %v", res.Status)
	}
}

func TestSentinel_TripsAfterThreshold(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("down")}}
	cfg := SentinelConfig{TripAfter: 2, ResetAfter: time.Minute, HealthyCode: StatusHealthy}
	c := NewSentinelChecker(inner, cfg)

	c.Check(context.Background(), "svc") // 1st — not yet tripped
	c.Check(context.Background(), "svc") // 2nd — trips
	res := c.Check(context.Background(), "svc") // should be suppressed
	if res.Status != StatusHealthy {
		t.Fatalf("expected sentinel to suppress, got %v", res.Status)
	}
	if inner.calls > 2 {
		t.Fatalf("inner should not be called while tripped, got %d calls", inner.calls)
	}
}

func TestSentinel_ResetsAfterDuration(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("down")}}
	cfg := SentinelConfig{TripAfter: 1, ResetAfter: 10 * time.Millisecond, HealthyCode: StatusHealthy}
	c := NewSentinelChecker(inner, cfg)

	c.Check(context.Background(), "svc") // trips
	time.Sleep(20 * time.Millisecond)

	// After reset, inner is called again.
	before := inner.calls
	c.Check(context.Background(), "svc")
	if inner.calls == before {
		t.Fatal("expected inner to be called after reset")
	}
}

func TestSentinel_DefaultConfigOnZero(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewSentinelChecker(inner, SentinelConfig{})
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("unexpected status %v", res.Status)
	}
}

func TestSentinel_DifferentServicesIndependent(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy, Err: errors.New("x")}}
	cfg := SentinelConfig{TripAfter: 2, ResetAfter: time.Minute, HealthyCode: StatusHealthy}
	c := NewSentinelChecker(inner, cfg)

	c.Check(context.Background(), "a")
	c.Check(context.Background(), "a") // a trips

	// b has not accumulated failures; inner should be called
	before := inner.calls
	c.Check(context.Background(), "b")
	if inner.calls == before {
		t.Fatal("b should delegate to inner independently")
	}
}
