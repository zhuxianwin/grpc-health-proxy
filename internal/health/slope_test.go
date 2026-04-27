package health

import (
	"context"
	"testing"
	"time"
)

func TestSlope_HealthyWhenAboveThreshold(t *testing.T) {
	// All healthy → slope ≈ 0, well above -0.3 threshold.
	inner := &stubChecker{result: Result{Status: StatusHealthy, CheckedAt: time.Now()}}
	c := NewSlopeChecker(inner, SlopeConfig{Window: 5, Threshold: -0.3}, nil)
	for i := 0; i < 5; i++ {
		res := c.Check(context.Background(), "svc")
		if res.Status != StatusHealthy {
			t.Fatalf("expected healthy, got %s", res.Status)
		}
	}
}

func TestSlope_UnhealthyWhenDeclining(t *testing.T) {
	// Alternate healthy/unhealthy in a clearly declining pattern.
	call := 0
	inner := CheckerFunc(func(_ context.Context, _ string) Result {
		call++
		// First half healthy, second half unhealthy → negative slope.
		if call <= 5 {
			return Result{Status: StatusHealthy, CheckedAt: time.Now()}
		}
		return Result{Status: StatusUnhealthy, CheckedAt: time.Now()}
	})
	c := NewSlopeChecker(inner, SlopeConfig{Window: 6, Threshold: -0.1}, nil)
	var last Result
	for i := 0; i < 10; i++ {
		last = c.Check(context.Background(), "svc")
	}
	if last.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy due to declining slope, got %s", last.Status)
	}
}

func TestSlope_DefaultConfigOnZero(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy, CheckedAt: time.Now()}}
	c := NewSlopeChecker(inner, SlopeConfig{}, nil)
	sc := c.(*slopeChecker)
	def := DefaultSlopeConfig()
	if sc.cfg.Window != def.Window {
		t.Fatalf("expected default window %d, got %d", def.Window, sc.cfg.Window)
	}
}

func TestSlope_DifferentServicesIndependent(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy, CheckedAt: time.Now()}}
	c := NewSlopeChecker(inner, SlopeConfig{Window: 3, Threshold: -0.1}, nil)
	for i := 0; i < 3; i++ {
		c.Check(context.Background(), "svc-a")
		c.Check(context.Background(), "svc-b")
	}
	sc := c.(*slopeChecker)
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if _, ok := sc.bucket["svc-a"]; !ok {
		t.Fatal("svc-a bucket missing")
	}
	if _, ok := sc.bucket["svc-b"]; !ok {
		t.Fatal("svc-b bucket missing")
	}
}

func TestComputeSlope_AllTrue(t *testing.T) {
	pts := []bool{true, true, true, true}
	s := computeSlope(pts)
	if s != 0 {
		t.Fatalf("expected slope 0 for constant series, got %f", s)
	}
}

func TestSlopeHandler_ReturnsJSON(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy, CheckedAt: time.Now()}}
	c := NewSlopeChecker(inner, DefaultSlopeConfig(), nil)
	h := NewSlopeStatusHandler(c)
	rr := serveHTTP(h, "GET", "/")
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}

func TestSlopeHandler_NotASlopeChecker(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy, CheckedAt: time.Now()}}
	h := NewSlopeStatusHandler(inner)
	rr := serveHTTP(h, "GET", "/")
	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
