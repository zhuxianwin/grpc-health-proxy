package health

import (
	"context"
	"testing"
	"time"
)

func TestReplay_RecordsHistory(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	rc := NewReplayChecker(inner, ReplayConfig{Window: 5 * time.Second, Capacity: 10})

	for i := 0; i < 3; i++ {
		_, _ = rc.Check(context.Background(), "svc")
	}

	h := rc.History("svc")
	if len(h) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(h))
	}
	for _, r := range h {
		if r.Status != StatusHealthy {
			t.Fatalf("expected healthy, got %s", r.Status)
		}
	}
}

func TestReplay_CapacityEviction(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	rc := NewReplayChecker(inner, ReplayConfig{Window: time.Minute, Capacity: 5})

	for i := 0; i < 10; i++ {
		_, _ = rc.Check(context.Background(), "svc")
	}

	h := rc.History("svc")
	if len(h) > 5 {
		t.Fatalf("expected at most 5 entries, got %d", len(h))
	}
}

func TestReplay_WindowEviction(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	rc := NewReplayChecker(inner, ReplayConfig{Window: 50 * time.Millisecond, Capacity: 100})

	_, _ = rc.Check(context.Background(), "svc")

	time.Sleep(80 * time.Millisecond)

	h := rc.History("svc")
	if len(h) != 0 {
		t.Fatalf("expected 0 entries after window expired, got %d", len(h))
	}
}

func TestReplay_DifferentServicesIndependent(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	rc := NewReplayChecker(inner, DefaultReplayConfig())

	_, _ = rc.Check(context.Background(), "alpha")
	_, _ = rc.Check(context.Background(), "alpha")
	_, _ = rc.Check(context.Background(), "beta")

	if len(rc.History("alpha")) != 2 {
		t.Fatal("expected 2 entries for alpha")
	}
	if len(rc.History("beta")) != 1 {
		t.Fatal("expected 1 entry for beta")
	}
}

func TestReplay_DefaultConfigOnZero(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	rc := NewReplayChecker(inner, ReplayConfig{})

	if rc.cfg.Window != DefaultReplayConfig().Window {
		t.Fatalf("expected default window, got %s", rc.cfg.Window)
	}
	if rc.cfg.Capacity != DefaultReplayConfig().Capacity {
		t.Fatalf("expected default capacity, got %d", rc.cfg.Capacity)
	}
}

func TestReplay_MissingServiceReturnsEmpty(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	rc := NewReplayChecker(inner, DefaultReplayConfig())

	h := rc.History("nonexistent")
	if len(h) != 0 {
		t.Fatalf("expected empty history, got %d entries", len(h))
	}
}
