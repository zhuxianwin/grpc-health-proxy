package health

import (
	"context"
	"testing"
	"time"
)

func TestObserve_CallbackInvoked(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	var got Result
	var gotService string
	var gotElapsed time.Duration

	cfg := ObserveConfig{
		OnResult: func(svc string, r Result, d time.Duration) {
			gotService = svc
			got = r
			gotElapsed = d
		},
	}

	c := NewObserveChecker(inner, cfg, nil)
	r := c.Check(context.Background(), "svc")

	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", r.Status)
	}
	if gotService != "svc" {
		t.Errorf("expected service svc, got %s", gotService)
	}
	if got.Status != StatusHealthy {
		t.Errorf("callback got wrong status")
	}
	if gotElapsed < 0 {
		t.Errorf("elapsed should be non-negative")
	}
}

func TestObserve_NilCallbackUsesDefault(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusUnhealthy}}
	c := NewObserveChecker(inner, ObserveConfig{}, nil)
	r := c.Check(context.Background(), "svc")
	if r.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy")
	}
}

func TestObserve_PropagatesError(t *testing.T) {
	inner := &fakeChecker{err: errUnreachable}
	var called bool
	cfg := ObserveConfig{
		OnResult: func(_ string, _ Result, _ time.Duration) { called = true },
	}
	c := NewObserveChecker(inner, cfg, nil)
	r := c.Check(context.Background(), "svc")
	if r.Err == nil {
		t.Fatal("expected error to propagate")
	}
	if !called {
		t.Fatal("callback should still be called on error")
	}
}

func TestDefaultObserveConfig(t *testing.T) {
	cfg := DefaultObserveConfig()
	if cfg.OnResult == nil {
		t.Fatal("OnResult should not be nil")
	}
	// should not panic
	cfg.OnResult("svc", Result{}, 0)
}
