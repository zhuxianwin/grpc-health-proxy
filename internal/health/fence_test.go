package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFence_DelegatesWhenLowered(t *testing.T) {
	inner := &staticChecker{result: Result{Status: StatusHealthy}}
	f := NewFenceChecker(inner, DefaultFenceConfig())

	res := f.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
}

func TestFence_BlocksWhenRaised(t *testing.T) {
	inner := &staticChecker{result: Result{Status: StatusHealthy}}
	f := NewFenceChecker(inner, DefaultFenceConfig())

	f.Raise()
	res := f.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy while fence raised, got %s", res.Status)
	}
	if !errors.Is(res.Err, ErrFenceRaised) {
		t.Fatalf("expected ErrFenceRaised, got %v", res.Err)
	}
}

func TestFence_LowerRestoresDelegation(t *testing.T) {
	inner := &staticChecker{result: Result{Status: StatusHealthy}}
	f := NewFenceChecker(inner, DefaultFenceConfig())

	f.Raise()
	f.Lower()
	res := f.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy after lower, got %s", res.Status)
	}
}

func TestFence_AutoExpires(t *testing.T) {
	inner := &staticChecker{result: Result{Status: StatusHealthy}}
	cfg := FenceConfig{OpenDuration: 20 * time.Millisecond}
	f := NewFenceChecker(inner, cfg)

	f.Raise()
	time.Sleep(40 * time.Millisecond)

	if f.IsRaised() {
		t.Fatal("expected fence to have auto-expired")
	}
	res := f.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy after expiry, got %s", res.Status)
	}
}

func TestFence_DefaultConfigOnZero(t *testing.T) {
	inner := &staticChecker{result: Result{Status: StatusHealthy}}
	f := NewFenceChecker(inner, FenceConfig{})
	if f.cfg.OpenDuration != DefaultFenceConfig().OpenDuration {
		t.Fatalf("expected default open duration, got %s", f.cfg.OpenDuration)
	}
}

func TestFenceHandler_ReturnsJSON(t *testing.T) {
	inner := &staticChecker{result: Result{Status: StatusHealthy}}
	f := NewFenceChecker(inner, DefaultFenceConfig())
	h := NewFenceStatusHandler(f)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "raised") {
		t.Fatalf("expected 'raised' key in response body")
	}
}

func TestFenceHandler_RaiseAndLowerViaHTTP(t *testing.T) {
	inner := &staticChecker{result: Result{Status: StatusHealthy}}
	f := NewFenceChecker(inner, DefaultFenceConfig())
	h := NewFenceStatusHandler(f)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/raise", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("raise: expected 200, got %d", rec.Code)
	}
	if !f.IsRaised() {
		t.Fatal("expected fence to be raised after POST /raise")
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/lower", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("lower: expected 200, got %d", rec.Code)
	}
	if f.IsRaised() {
		t.Fatal("expected fence to be lowered after POST /lower")
	}
}

func TestFenceHandler_NotAFenceChecker(t *testing.T) {
	inner := &staticChecker{result: Result{Status: StatusHealthy}}
	h := NewFenceStatusHandler(inner)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-FenceChecker, got %d", rec.Code)
	}
}
