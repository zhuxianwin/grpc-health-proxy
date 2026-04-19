package health_test

import (
	"context"
	"testing"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

func TestMirror_Integration_PrimaryWins(t *testing.T) {
	healthy := health.NewFallbackChecker(
		&staticChecker{status: health.StatusHealthy},
		nil,
		health.DefaultFallbackConfig(),
	)
	unhealthy := health.NewFallbackChecker(
		&staticChecker{status: health.StatusUnhealthy},
		nil,
		health.DefaultFallbackConfig(),
	)

	mc := health.NewMirrorChecker(healthy, []health.Checker{unhealthy}, health.DefaultMirrorConfig())
	res := mc.Check(context.Background(), "integration-svc")

	if res.Status != health.StatusHealthy {
		t.Fatalf("expected healthy from primary, got %v", res.Status)
	}
}

// staticChecker is a minimal Checker for integration tests.
type staticChecker struct{ status health.Status }

func (s *staticChecker) Check(_ context.Context, _ string) health.Result {
	return health.Result{Status: s.status}
}
