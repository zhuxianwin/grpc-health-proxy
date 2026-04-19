package health

import (
	"context"
	"testing"
)

func TestFlip_DelegatesWhenInactive(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	fc := NewFlipChecker(inner, DefaultFlipConfig())
	res := fc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
}

func TestFlip_InvertsHealthyToUnhealthy(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	fc := NewFlipChecker(inner, DefaultFlipConfig())
	fc.Toggle()
	res := fc.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", res.Status)
	}
}

func TestFlip_InvertsUnhealthyToHealthy(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	fc := NewFlipChecker(inner, DefaultFlipConfig())
	fc.Toggle()
	res := fc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", res.Status)
	}
}

func TestFlip_ToggleRestoresBehaviour(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	fc := NewFlipChecker(inner, DefaultFlipConfig())
	fc.Toggle() // on
	fc.Toggle() // off
	res := fc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatalf("expected healthy after double toggle, got %s", res.Status)
	}
}

func TestFlip_ErrorPassedThrough(t *testing.T) {
	errSentinel := errUnreachable
	inner := &stubChecker{result: Result{Err: errSentinel}}
	fc := NewFlipChecker(inner, DefaultFlipConfig())
	fc.Toggle()
	res := fc.Check(context.Background(), "svc")
	if res.Err != errSentinel {
		t.Fatalf("expected error to pass through, got %v", res.Err)
	}
}

func TestFlipHandler_ReturnsJSON(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	fc := NewFlipChecker(inner, DefaultFlipConfig())
	h := NewFlipStatusHandler(fc)

	rec, req := newRecorder(), newGetRequest("/")
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}
