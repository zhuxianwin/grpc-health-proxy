package health

import (
	"context"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type countingChecker struct {
	calls atomic.Int64
	result Result
}

func (c *countingChecker) Check(_ context.Context, _ string) Result {
	c.calls.Add(1)
	return c.result
}

func TestShadow_PrimaryResultReturned(t *testing.T) {
	primary := &countingChecker{result: Result{Status: StatusHealthy}}
	shadow := &countingChecker{result: Result{Status: StatusUnhealthy}}

	sc := NewShadowChecker(primary, shadow, DefaultShadowConfig(), nil)
	res := sc.Check(context.Background(), "svc")

	if res.Status != StatusHealthy {
		t.Fatalf("expected primary result, got %s", res.Status)
	}
}

func TestShadow_ShadowFiredOnSample(t *testing.T) {
	primary := &countingChecker{result: Result{Status: StatusHealthy}}
	shadow := &countingChecker{result: Result{Status: StatusHealthy}}

	cfg := ShadowConfig{SampleRate: 1.0, Timeout: 500 * time.Millisecond}
	sc := NewShadowChecker(primary, shadow, cfg, nil)

	sc.Check(context.Background(), "svc")
	time.Sleep(50 * time.Millisecond) // let goroutine run

	if shadow.calls.Load() == 0 {
		t.Fatal("expected shadow checker to be called")
	}
}

func TestShadow_ZeroRateNeverFires(t *testing.T) {
	primary := &countingChecker{result: Result{Status: StatusHealthy}}
	shadow := &countingChecker{result: Result{Status: StatusHealthy}}

	cfg := ShadowConfig{SampleRate: 0.0, Timeout: 500 * time.Millisecond}
	sc := NewShadowChecker(primary, shadow, cfg, nil)

	for i := 0; i < 100; i++ {
		sc.Check(context.Background(), "svc")
	}
	time.Sleep(50 * time.Millisecond)

	if shadow.calls.Load() != 0 {
		t.Fatalf("expected no shadow calls, got %d", shadow.calls.Load())
	}
}

func TestShadowHandler_ReturnsJSON(t *testing.T) {
	primary := &countingChecker{result: Result{Status: StatusHealthy}}
	shadow := &countingChecker{result: Result{Status: StatusHealthy}}
	cfg := DefaultShadowConfig()
	sc := NewShadowChecker(primary, shadow, cfg, nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/shadow", nil)
	NewShadowStatusHandler(sc).ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "sample_rate") {
		t.Fatal("expected sample_rate in response")
	}
}

func TestShadowHandler_NotAShadowChecker(t *testing.T) {
	c := &countingChecker{result: Result{Status: StatusHealthy}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/shadow", nil)
	NewShadowStatusHandler(c).ServeHTTP(rr, req)

	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
