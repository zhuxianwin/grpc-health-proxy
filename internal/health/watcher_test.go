package health_test

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/grpc-health-proxy/internal/health"
	"go.uber.org/zap"
)

func TestWatcher_PopulatesCache(t *testing.T) {
	addr, stop := startFakeHealthServer(t, true)
	defer stop()

	log, _ := zap.NewDevelopment()
	cache := health.NewCache(5 * time.Second)
	checker := health.NewChecker(addr, log)

	watcher := health.NewWatcher(checker, cache, []string{""}, 50*time.Millisecond, log)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	watcher.Start(ctx)
	<-ctx.Done()
	watcher.Stop()

	result, ok := cache.Get("")
	if !ok {
		t.Fatal("expected cache to contain a result after watcher ran")
	}
	if result.Status != health.StatusHealthy {
		t.Errorf("expected healthy, got %s", result.Status)
	}
}

func TestWatcher_UnhealthyService(t *testing.T) {
	addr, stop := startFakeHealthServer(t, false)
	defer stop()

	log, _ := zap.NewDevelopment()
	cache := health.NewCache(5 * time.Second)
	checker := health.NewChecker(addr, log)

	watcher := health.NewWatcher(checker, cache, []string{""}, 50*time.Millisecond, log)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	watcher.Start(ctx)
	<-ctx.Done()
	watcher.Stop()

	result, ok := cache.Get("")
	if !ok {
		t.Fatal("expected cache to contain a result after watcher ran")
	}
	if result.Status != health.StatusUnhealthy {
		t.Errorf("expected unhealthy, got %s", result.Status)
	}
}

func TestWatcher_StopsCleanly(t *testing.T) {
	addr, stop := startFakeHealthServer(t, true)
	defer stop()

	log := zap.NewNop()
	cache := health.NewCache(5 * time.Second)
	checker := health.NewChecker(addr, log)
	watcher := health.NewWatcher(checker, cache, []string{"svc-a", "svc-b"}, 10*time.Millisecond, log)

	ctx, cancel := context.WithCancel(context.Background())
	watcher.Start(ctx)

	time.Sleep(30 * time.Millisecond)
	cancel()

	done := make(chan struct{})
	go func() { watcher.Stop(); close(done) }()

	select {
	case <-done:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("watcher did not stop within timeout")
	}
}
