package health

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestBulkhead_AllowsUnderLimit(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	bc := NewBulkheadChecker(inner, BulkheadConfig{MaxConcurrent: 5})

	result := bc.Check(context.Background(), "svc")
	if result.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", result.Status)
	}
}

func TestBulkhead_RejectsOverLimit(t *testing.T) {
	blocked := make(chan struct{})
	release := make(chan struct{})

	slow := CheckerFunc(func(ctx context.Context, service string) Result {
		blocked <- struct{}{}
		<-release
		return Result{Service: service, Status: StatusHealthy}
	})

	bc := NewBulkheadChecker(slow, BulkheadConfig{MaxConcurrent: 2})

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bc.Check(context.Background(), "svc")
		}()
	}

	// wait for both goroutines to be inside the inner checker
	<-blocked
	<-blocked

	// now the bulkhead should be full
	result := bc.Check(context.Background(), "svc")
	if result.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy (full), got %s", result.Status)
	}
	if result.Err != ErrBulkheadFull {
		t.Fatalf("expected ErrBulkheadFull, got %v", result.Err)
	}

	close(release)
	wg.Wait()
}

func TestBulkhead_DefaultConfig(t *testing.T) {
	cfg := DefaultBulkheadConfig()
	if cfg.MaxConcurrent != 10 {
		t.Fatalf("expected default MaxConcurrent=10, got %d", cfg.MaxConcurrent)
	}
}

func TestBulkhead_ZeroMaxUsesDefault(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	bc := NewBulkheadChecker(inner, BulkheadConfig{MaxConcurrent: 0})

	result := bc.Check(context.Background(), "svc")
	if result.Status != StatusHealthy {
		t.Fatalf("expected healthy, got %s", result.Status)
	}
}

func TestBulkhead_ReleasesSlotAfterCheck(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	bc := NewBulkheadChecker(inner, BulkheadConfig{MaxConcurrent: 1})

	for i := 0; i < 5; i++ {
		r := bc.Check(context.Background(), "svc")
		if r.Status != StatusHealthy {
			t.Fatalf("iteration %d: expected healthy, got %s", i, r.Status)
		}
	}
}

// fakeChecker is a simple test double; reused across health package tests.
type fakeChecker struct {
	result Result
	delay  time.Duration
}

func (f *fakeChecker) Check(ctx context.Context, service string) Result {
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
	r := f.result
	r.Service = service
	return r
}
