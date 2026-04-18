package health

import (
	"context"
	"testing"
	"time"
)

func TestDrain_DelegatesWhenNotDraining(t *testing.T) {
	inner := &staticChecker{result: Result{Status: StatusHealthy}}
	d := NewDrainChecker(inner, DefaultDrainConfig(), nil)

	r := d.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", r.Status)
	}
}

func TestDrain_UnhealthyWhenDraining(t *testing.T) {
	inner := &staticChecker{result: Result{Status: StatusHealthy}}
	d := NewDrainChecker(inner, DefaultDrainConfig(), nil)

	d.draining.Store(true)
	r := d.Check(context.Background(), "svc")
	if r.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy during drain, got %s", r.Status)
	}
}

func TestDrain_DrainBlocksForGracePeriod(t *testing.T) {
	inner := &staticChecker{result: Result{Status: StatusHealthy}}
	cfg := DrainConfig{GracePeriod: 50 * time.Millisecond}
	d := NewDrainChecker(inner, cfg, nil)

	start := time.Now()
	d.Drain(context.Background())
	elapsed := time.Since(start)

	if elapsed < 40*time.Millisecond {
		t.Fatalf("drain returned too quickly: %s", elapsed)
	}
	if !d.Draining() {
		t.Fatal("expected Draining() == true after Drain")
	}
}

func TestDrain_DrainRespectsContextCancel(t *testing.T) {
	inner := &staticChecker{result: Result{Status: StatusHealthy}}
	cfg := DrainConfig{GracePeriod: 10 * time.Second}
	d := NewDrainChecker(inner, cfg, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	start := time.Now()
	d.Drain(ctx)
	if time.Since(start) > 500*time.Millisecond {
		t.Fatal("drain did not respect context cancellation")
	}
}

func TestDrain_DefaultConfigOnZero(t *testing.T) {
	d := NewDrainChecker(&staticChecker{}, DrainConfig{}, nil)
	if d.config.GracePeriod != DefaultDrainConfig().GracePeriod {
		t.Fatalf("expected default grace period, got %s", d.config.GracePeriod)
	}
}
