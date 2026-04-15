package health

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestAdmission_AllowsUnderLimit(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	checker := NewAdmissionChecker(inner, AdmissionConfig{MaxConcurrent: 5}, nil)

	result := checker.Check(context.Background(), "svc")
	if result.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", result.Status)
	}
}

func TestAdmission_RejectsOverLimit(t *testing.T) {
	blocked := make(chan struct{})
	release := make(chan struct{})

	inner := CheckerFunc(func(ctx context.Context, service string) Result {
		close(blocked)
		<-release
		return Result{Status: StatusHealthy}
	})

	checker := NewAdmissionChecker(inner, AdmissionConfig{MaxConcurrent: 1}, nil)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		checker.Check(context.Background(), "svc")
	}()

	select {
		case <-blocked:
		case <-time.After(time.Second):
			t.Fatal("goroutine did not start")
	}

	result := checker.Check(context.Background(), "svc")
	if result.Err != ErrAdmissionLimitExceeded {
		t.Fatalf("expected ErrAdmissionLimitExceeded, got %v", result.Err)
	}
	if result.Status != StatusUnknown {
		t.Fatalf("expected unknown status, got %s", result.Status)
	}

	close(release)
	wg.Wait()
}

func TestAdmission_ReleasesSlotAfterCheck(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	checker := NewAdmissionChecker(inner, AdmissionConfig{MaxConcurrent: 1}, nil)

	for i := 0; i < 3; i++ {
		result := checker.Check(context.Background(), "svc")
		if result.Status != StatusHealthy {
			t.Fatalf("iteration %d: expected healthy, got %s", i, result.Status)
		}
	}
}

func TestAdmission_DefaultConfig(t *testing.T) {
	cfg := DefaultAdmissionConfig()
	if cfg.MaxConcurrent != 10 {
		t.Fatalf("expected default MaxConcurrent=10, got %d", cfg.MaxConcurrent)
	}
}

func TestAdmission_ZeroMaxUsesDefault(t *testing.T) {
	inner := &stubChecker{result: Result{Status: StatusHealthy}}
	checker := NewAdmissionChecker(inner, AdmissionConfig{MaxConcurrent: 0}, nil)

	result := checker.Check(context.Background(), "svc")
	if result.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", result.Status)
	}
}
