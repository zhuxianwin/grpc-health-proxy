package health

import (
	"context"
	"testing"
	"time"
)

func TestCooldown_HealthyPassesThrough(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	cc := NewCooldownChecker(inner, DefaultCooldownConfig(), nil)
	res, err := cc.Check(context.Background(), "svc")
	if err != nil || res.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %v %v", res.Status, err)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls)
	}
}

func TestCooldown_UnhealthyTriggersCooldown(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	cfg := CooldownConfig{Duration: 200 * time.Millisecond}
	cc := NewCooldownChecker(inner, cfg, nil)

	res, _ := cc.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatal("first call should be unhealthy")
	}

	// second call should be served from cooldown cache
	inner.result = Result{Status: StatusHealthy}
	res, _ = cc.Check(context.Background(), "svc")
	if res.Status != StatusUnhealthy {
		t.Fatal("expected cached unhealthy during cooldown")
	}
	if inner.calls != 1 {
		t.Fatalf("inner should only be called once, got %d", inner.calls)
	}
}

func TestCooldown_ExpiresAfterDuration(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	cfg := CooldownConfig{Duration: 30 * time.Millisecond}
	cc := NewCooldownChecker(inner, cfg, nil)

	cc.Check(context.Background(), "svc") //nolint
	time.Sleep(50 * time.Millisecond)

	inner.result = Result{Status: StatusHealthy}
	res, _ := cc.Check(context.Background(), "svc")
	if res.Status != StatusHealthy {
		t.Fatal("expected healthy after cooldown expiry")
	}
}

func TestCooldown_DefaultConfigOnZero(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	cc := NewCooldownChecker(inner, CooldownConfig{}, nil)
	if cc.cfg.Duration != DefaultCooldownConfig().Duration {
		t.Fatalf("expected default duration, got %v", cc.cfg.Duration)
	}
}

func TestCooldown_DifferentServicesIndependent(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusUnhealthy}}
	cfg := CooldownConfig{Duration: 200 * time.Millisecond}
	cc := NewCooldownChecker(inner, cfg, nil)

	cc.Check(context.Background(), "svc-a") //nolint

	inner.result = Result{Status: StatusHealthy}
	res, _ := cc.Check(context.Background(), "svc-b")
	if res.Status != StatusHealthy {
		t.Fatal("svc-b should not be affected by svc-a cooldown")
	}
}
