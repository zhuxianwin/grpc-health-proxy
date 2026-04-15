package health_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/grpc-health-proxy/internal/health"
)

func TestMultiChecker_AllHealthy(t *testing.T) {
	addr1 := startFakeHealthServer(t, true)
	addr2 := startFakeHealthServer(t, true)

	cache := health.NewCache(5 * time.Second)
	mc := health.NewMultiChecker([]string{addr1, addr2}, cache)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	results := mc.CheckAll(ctx, "")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !mc.Healthy(results) {
		t.Error("expected all targets to be healthy")
	}
}

func TestMultiChecker_OneUnhealthy(t *testing.T) {
	addr1 := startFakeHealthServer(t, true)
	addr2 := startFakeHealthServer(t, false)

	cache := health.NewCache(5 * time.Second)
	mc := health.NewMultiChecker([]string{addr1, addr2}, cache)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	results := mc.CheckAll(ctx, "")
	if mc.Healthy(results) {
		t.Error("expected overall unhealthy when one target is down")
	}
}

func TestMultiChecker_EmptyTargets(t *testing.T) {
	cache := health.NewCache(5 * time.Second)
	mc := health.NewMultiChecker([]string{}, cache)

	ctx := context.Background()
	results := mc.CheckAll(ctx, "")
	if len(results) != 0 {
		t.Fatalf("expected 0 results for empty targets")
	}
	if !mc.Healthy(results) {
		t.Error("empty target set should be considered healthy")
	}
}
