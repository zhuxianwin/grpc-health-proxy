package health

import (
	"context"
	"testing"
	"time"
)

func TestQuota_AllowsUnderLimit(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	qc := NewQuotaChecker(inner, QuotaConfig{MaxCalls: 5, Window: time.Minute}, nil)

	for i := 0; i < 5; i++ {
		r := qc.Check(context.Background(), "svc")
		if r.Err != nil {
			t.Fatalf("unexpected error on call %d: %v", i+1, r.Err)
		}
	}
}

func TestQuota_RejectsOverLimit(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	qc := NewQuotaChecker(inner, QuotaConfig{MaxCalls: 2, Window: time.Minute}, nil)

	qc.Check(context.Background(), "svc")
	qc.Check(context.Background(), "svc")
	r := qc.Check(context.Background(), "svc")
	if r.Err == nil {
		t.Fatal("expected quota error, got nil")
	}
}

func TestQuota_DefaultConfig(t *testing.T) {
	cfg := DefaultQuotaConfig()
	if cfg.MaxCalls <= 0 {
		t.Error("MaxCalls should be positive")
	}
	if cfg.Window <= 0 {
		t.Error("Window should be positive")
	}
}

func TestQuota_ZeroMaxUsesDefault(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	qc := NewQuotaChecker(inner, QuotaConfig{MaxCalls: 0, Window: time.Minute}, nil)
	r := qc.Check(context.Background(), "svc")
	if r.Err != nil {
		t.Fatalf("first call should succeed with default max: %v", r.Err)
	}
}

func TestQuota_NilLogger(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	qc := NewQuotaChecker(inner, DefaultQuotaConfig(), nil)
	r := qc.Check(context.Background(), "svc")
	if r.Err != nil {
		t.Fatalf("unexpected error: %v", r.Err)
	}
}
