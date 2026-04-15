package health_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourorg/grpc-health-proxy/internal/health"
)

// slowChecker simulates a Checker that blocks until its context is cancelled
// or the provided delay elapses.
type slowChecker struct {
	delay  time.Duration
	status health.Status
}

func (s *slowChecker) Check(ctx context.Context, service string) health.Result {
	select {
	case <-time.After(s.delay):
		return health.Result{Service: service, Status: s.status}
	case <-ctx.Done():
		return health.Result{Service: service, Status: health.StatusUnknown, Err: ctx.Err()}
	}
}

func TestTimeoutChecker_CompletesInTime(t *testing.T) {
	inner := &slowChecker{delay: 10 * time.Millisecond, status: health.StatusHealthy}
	tc := health.NewTimeoutChecker(inner, 500*time.Millisecond)

	res := tc.Check(context.Background(), "svc")
	if res.Status != health.StatusHealthy {
		t.Fatalf("expected Healthy, got %s", res.Status)
	}
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
}

func TestTimeoutChecker_Expires(t *testing.T) {
	inner := &slowChecker{delay: 500 * time.Millisecond, status: health.StatusHealthy}
	tc := health.NewTimeoutChecker(inner, 30*time.Millisecond)

	res := tc.Check(context.Background(), "svc")
	if res.Status != health.StatusUnknown {
		t.Fatalf("expected Unknown on timeout, got %s", res.Status)
	}
	if res.Err == nil {
		t.Fatal("expected non-nil error on timeout")
	}
}

func TestTimeoutChecker_DefaultTimeout(t *testing.T) {
	// zero timeout should fall back to 5 s default (just verify no panic / sane result)
	inner := &slowChecker{delay: 5 * time.Millisecond, status: health.StatusHealthy}
	tc := health.NewTimeoutChecker(inner, 0)

	res := tc.Check(context.Background(), "svc")
	if res.Status != health.StatusHealthy {
		t.Fatalf("expected Healthy, got %s", res.Status)
	}
}

func TestTimeoutChecker_PropagatesUnhealthy(t *testing.T) {
	inner := &slowChecker{delay: 5 * time.Millisecond, status: health.StatusUnhealthy}
	tc := health.NewTimeoutChecker(inner, 500*time.Millisecond)

	res := tc.Check(context.Background(), "svc")
	if res.Status != health.StatusUnhealthy {
		t.Fatalf("expected Unhealthy, got %s", res.Status)
	}
}

func TestTimeoutChecker_ParentCancellation(t *testing.T) {
	inner := &slowChecker{delay: 500 * time.Millisecond, status: health.StatusHealthy}
	tc := health.NewTimeoutChecker(inner, 5*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	res := tc.Check(ctx, "svc")
	if res.Err == nil {
		t.Fatal("expected error when parent context is cancelled")
	}
	if !errors.Is(res.Err, context.Canceled) {
		t.Logf("got error (acceptable): %v", res.Err)
	}
}
