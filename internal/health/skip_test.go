package health

import (
	"context"
	"errors"
	"testing"
)

func TestSkip_DelegatesWhenNotSkipping(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	s := NewSkipChecker(inner, DefaultSkipConfig())

	r := s.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", r.Status)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", inner.calls)
	}
}

func TestSkip_BypassesInnerWhenSkipping(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	s := NewSkipChecker(inner, DefaultSkipConfig())

	// Prime a healthy result so last is populated.
	inner.result = Result{Status: StatusHealthy}
	s.Check(context.Background(), "svc")

	// Now flip to unhealthy and enable skip.
	inner.result = Result{Status: StatusUnhealthy}
	s.Skip()

	r := s.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected cached healthy result, got %s", r.Status)
	}
	if inner.calls != 1 {
		t.Fatalf("inner should not have been called again, got %d calls", inner.calls)
	}
}

func TestSkip_ReturnsHealthyWhenNoCachedResult(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	s := NewSkipChecker(inner, DefaultSkipConfig())
	s.Skip()

	r := s.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected synthesised healthy, got %s", r.Status)
	}
	if inner.calls != 0 {
		t.Fatalf("inner should not have been called, got %d calls", inner.calls)
	}
}

func TestSkip_ResumeRestoresDelegation(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	s := NewSkipChecker(inner, DefaultSkipConfig())
	s.Skip()
	s.Resume()

	r := s.Check(context.Background(), "svc")
	if r.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy after resume, got %s", r.Status)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 inner call after resume, got %d", inner.calls)
	}
}

func TestSkip_PropagatesError(t *testing.T) {
	sentinel := errors.New("dial error")
	inner := &stubChecker{result: Result{Status: StatusUnhealthy, Err: sentinel}}
	s := NewSkipChecker(inner, DefaultSkipConfig())

	r := s.Check(context.Background(), "svc")
	if !errors.Is(r.Err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", r.Err)
	}
}

func TestSkip_DefaultConfigOnZero(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	s := NewSkipChecker(inner, SkipConfig{})
	if s.cfg.Logger == nil {
		t.Fatal("expected non-nil logger from zero config")
	}
}
