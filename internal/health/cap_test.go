package health

import (
	"context"
	"testing"
)

func TestCap_BelowCapPassesThrough(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	c := NewCapChecker(inner, CapConfig{Max: StatusHealthy})
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected %v, got %v", StatusHealthy, res.Status)
	}
}

func TestCap_AboveCapIsClamped(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	c := NewCapChecker(inner, CapConfig{Max: StatusHealthy})
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected status clamped to %v, got %v", StatusHealthy, res.Status)
	}
}

func TestCap_EqualToCapPassesThrough(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	c := NewCapChecker(inner, CapConfig{Max: StatusUnhealthy})
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected %v, got %v", StatusUnhealthy, res.Status)
	}
}

func TestCap_DefaultConfigOnZero(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	// Zero CapConfig should trigger DefaultCapConfig (Max = StatusHealthy).
	c := NewCapChecker(inner, CapConfig{})
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected %v, got %v", StatusHealthy, res.Status)
	}
}

func TestCap_NilLoggerUsesDefault(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	c := NewCapChecker(inner, CapConfig{Max: StatusHealthy, Logger: nil})
	// Should not panic.
	res := c.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected %v, got %v", StatusHealthy, res.Status)
	}
}
