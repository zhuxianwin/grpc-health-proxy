package health

import (
	"context"
	"testing"
)

func TestWindow_HealthyAboveThreshold(t *testing.T) {
	inner := &stubChecker{status: StatusHealthy}
	w := NewWindowChecker(inner, WindowConfig{Size: 5, Threshold: 0.6})

	// Fill window with healthy results.
	for i := 0; i < 5; i++ {
		res := w.Check(context.Background(), "svc")
		if res.Status != StatusHealthy {
			t.Fatalf("expected healthy, got %s", res.Status)
		}
	}
}

func TestWindow_UnhealthyBelowThreshold(t *testing.T) {
	// Alternate unhealthy/healthy so fraction ends up below 0.6.
	call := 0
	inner := &fnChecker{fn: func(service string) Result {
		call++
		if call%2 == 0 {
			return Result{Service: service, Status: StatusHealthy}
		}
		return Result{Service: service, Status: StatusUnhealthy}
	}}
	w := NewWindowChecker(inner, WindowConfig{Size: 4, Threshold: 0.8})

	var last Result
	for i := 0; i < 4; i++ {
		last = w.Check(context.Background(), "svc")
	}
	// 2 healthy out of 4 = 0.5, below 0.8 threshold.
	if last.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy due to window, got %s", last.Status)
	}
}

func TestWindow_DefaultConfigOnZero(t *testing.T) {
	inner := &stubChecker{status: StatusHealthy}
	w := NewWindowChecker(inner, WindowConfig{})
	if w.cfg.Size != DefaultWindowConfig().Size {
		t.Fatalf("expected default size %d, got %d", DefaultWindowConfig().Size, w.cfg.Size)
	}
	if w.cfg.Threshold != DefaultWindowConfig().Threshold {
		t.Fatalf("expected default threshold %f, got %f", DefaultWindowConfig().Threshold, w.cfg.Threshold)
	}
}

func TestWindow_SlidingEviction(t *testing.T) {
	call := 0
	inner := &fnChecker{fn: func(service string) Result {
		call++
		// First 3 unhealthy, then all healthy.
		if call <= 3 {
			return Result{Service: service, Status: StatusUnhealthy}
		}
		return Result{Service: service, Status: StatusHealthy}
	}}
	w := NewWindowChecker(inner, WindowConfig{Size: 3, Threshold: 0.6})

	for i := 0; i < 6; i++ {
		w.Check(context.Background(), "svc")
	}
	_, frac := w.WindowStats()
	if frac < 0.6 {
		t.Fatalf("expected window fraction >= 0.6 after eviction, got %f", frac)
	}
}

func TestWindow_StatsEmptyWindow(t *testing.T) {
	w := NewWindowChecker(&stubChecker{status: StatusHealthy}, WindowConfig{})
	size, frac := w.WindowStats()
	if size != 0 || frac != 0 {
		t.Fatalf("expected 0,0 for empty window, got %d,%f", size, frac)
	}
}

// fnChecker calls fn for each Check.
type fnChecker struct{ fn func(string) Result }

func (f *fnChecker) Check(_ context.Context, service string) Result { return f.fn(service) }
