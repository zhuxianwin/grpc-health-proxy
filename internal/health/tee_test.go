package health_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/grpc-health-proxy/internal/health"
)

type recordingChecker struct {
	service string
	called  atomic.Bool
	result  health.Result
	err     error
}

func (r *recordingChecker) Check(ctx context.Context, service string) (health.Result, error) {
	r.service = service
	r.called.Store(true)
	return r.result, r.err
}

func TestTee_PrimaryResultReturned(t *testing.T) {
	primary := &fakeChecker{result: health.Result{Status: health.StatusHealthy}}
	secondary := &recordingChecker{result: health.Result{Status: health.StatusHealthy}}

	cfg := health.DefaultTeeConfig()
	checker := health.NewTeeChecker(cfg, primary, secondary)

	res, err := checker.Check(context.Background(), "svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != health.StatusHealthy {
		t.Errorf("expected healthy, got %s", res.Status)
	}
}

func TestTee_SecondaryReceivesSameService(t *testing.T) {
	primary := &fakeChecker{result: health.Result{Status: health.StatusHealthy}}
	secondary := &recordingChecker{result: health.Result{Status: health.StatusHealthy}}

	cfg := health.DefaultTeeConfig()
	checker := health.NewTeeChecker(cfg, primary, secondary)

	_, _ = checker.Check(context.Background(), "my-service")

	// Allow async secondary to complete.
	time.Sleep(20 * time.Millisecond)

	if !secondary.called.Load() {
		t.Error("secondary checker was not called")
	}
	if secondary.service != "my-service" {
		t.Errorf("expected service %q, got %q", "my-service", secondary.service)
	}
}

func TestTee_SecondaryErrorDoesNotPropagate(t *testing.T) {
	primary := &fakeChecker{result: health.Result{Status: health.StatusHealthy}}
	secondary := &recordingChecker{err: errors.New("secondary boom")}

	cfg := health.DefaultTeeConfig()
	checker := health.NewTeeChecker(cfg, primary, secondary)

	res, err := checker.Check(context.Background(), "svc")
	if err != nil {
		t.Fatalf("secondary error should not propagate, got: %v", err)
	}
	if res.Status != health.StatusHealthy {
		t.Errorf("expected healthy, got %s", res.Status)
	}
}

func TestTee_PrimaryErrorPropagates(t *testing.T) {
	primary := &fakeChecker{err: errors.New("primary boom")}
	secondary := &recordingChecker{result: health.Result{Status: health.StatusHealthy}}

	cfg := health.DefaultTeeConfig()
	checker := health.NewTeeChecker(cfg, primary, secondary)

	_, err := checker.Check(context.Background(), "svc")
	if err == nil {
		t.Fatal("expected primary error to propagate")
	}
}

func TestTee_DefaultConfigOnZero(t *testing.T) {
	cfg := health.DefaultTeeConfig()
	if cfg.Timeout <= 0 {
		t.Errorf("expected positive default timeout, got %v", cfg.Timeout)
	}
}

func TestTee_NoSecondaries(t *testing.T) {
	primary := &fakeChecker{result: health.Result{Status: health.StatusHealthy}}

	cfg := health.DefaultTeeConfig()
	checker := health.NewTeeChecker(cfg, primary)

	res, err := checker.Check(context.Background(), "svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != health.StatusHealthy {
		t.Errorf("expected healthy, got %s", res.Status)
	}
}
