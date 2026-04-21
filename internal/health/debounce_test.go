package health

import (
	"context"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type countingChecker struct {
	calls atomic.Int32
	res   Result
	err   error
}

func (c *countingChecker) Check(_ context.Context, _ string) (Result, error) {
	c.calls.Add(1)
	return c.res, c.err
}

func TestDebounce_FirstCallHitsInner(t *testing.T) {
	inner := &countingChecker{res: Healthy("svc")}
	d := NewDebounceChecker(inner, DefaultDebounceConfig(), nil)

	_, err := d.Check(context.Background(), "svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls.Load())
	}
}

func TestDebounce_SecondCallWithinWindowIsCached(t *testing.T) {
	inner := &countingChecker{res: Healthy("svc")}
	cfg := DebounceConfig{Window: 10 * time.Second}
	d := NewDebounceChecker(inner, cfg, nil)

	d.Check(context.Background(), "svc") //nolint:errcheck
	d.Check(context.Background(), "svc") //nolint:errcheck

	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 upstream call, got %d", inner.calls.Load())
	}
}

func TestDebounce_CallAfterWindowHitsInner(t *testing.T) {
	inner := &countingChecker{res: Healthy("svc")}
	cfg := DebounceConfig{Window: 10 * time.Millisecond}
	d := NewDebounceChecker(inner, cfg, nil)

	d.Check(context.Background(), "svc") //nolint:errcheck
	time.Sleep(20 * time.Millisecond)
	d.Check(context.Background(), "svc") //nolint:errcheck

	if inner.calls.Load() != 2 {
		t.Fatalf("expected 2 upstream calls, got %d", inner.calls.Load())
	}
}

func TestDebounce_DefaultConfigOnZero(t *testing.T) {
	inner := &countingChecker{res: Healthy("svc")}
	d := NewDebounceChecker(inner, DebounceConfig{}, nil).(*debounceChecker)
	if d.cfg.Window != DefaultDebounceConfig().Window {
		t.Fatalf("expected default window %v, got %v", DefaultDebounceConfig().Window, d.cfg.Window)
	}
}

func TestDebounceHandler_ReturnsJSON(t *testing.T) {
	inner := &countingChecker{res: Healthy("svc")}
	d := NewDebounceChecker(inner, DefaultDebounceConfig(), nil)
	h := NewDebounceStatusHandler(d)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}

func TestDebounceHandler_NotADebounceChecker(t *testing.T) {
	h := NewDebounceStatusHandler(&countingChecker{})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
