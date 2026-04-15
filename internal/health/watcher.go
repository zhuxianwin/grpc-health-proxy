package health

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Watcher periodically polls a set of gRPC services and refreshes the cache.
type Watcher struct {
	checker  *Checker
	cache    *Cache
	services []string
	interval time.Duration
	log      *zap.Logger
	wg       sync.WaitGroup
}

// NewWatcher creates a Watcher that polls the given services at the specified interval.
func NewWatcher(checker *Checker, cache *Cache, services []string, interval time.Duration, log *zap.Logger) *Watcher {
	return &Watcher{
		checker:  checker,
		cache:    cache,
		services: services,
		interval: interval,
		log:      log,
	}
}

// Start begins background polling until ctx is cancelled.
func (w *Watcher) Start(ctx context.Context) {
	for _, svc := range w.services {
		svc := svc
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			w.poll(ctx, svc)
		}()
	}
}

// Stop waits for all polling goroutines to finish.
func (w *Watcher) Stop() {
	w.wg.Wait()
}

func (w *Watcher) poll(ctx context.Context, service string) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Perform an immediate check before waiting for the first tick.
	w.check(ctx, service)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.check(ctx, service)
		}
	}
}

func (w *Watcher) check(ctx context.Context, service string) {
	result := w.checker.Check(ctx, service)
	w.cache.Set(service, result)
	w.log.Debug("watcher refreshed service",
		zap.String("service", service),
		zap.String("status", result.Status.String()),
	)
}
