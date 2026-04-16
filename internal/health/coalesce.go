package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// CoalesceConfig controls the coalescing window.
type CoalesceConfig struct {
	// Window is how long to batch incoming requests before firing one check.
	Window time.Duration
}

// DefaultCoalesceConfig returns sensible defaults.
func DefaultCoalesceConfig() CoalesceConfig {
	return CoalesceConfig{Window: 50 * time.Millisecond}
}

type coalesceChecker struct {
	inner  Checker
	cfg    CoalesceConfig
	log    *slog.Logger

	mu      sync.Mutex
	flights map[string]*coalesceFlight
}

type coalesceFlight struct {
	doneCh chan struct{}
	result Result
}

// NewCoalesceChecker wraps a Checker so that concurrent calls for the same
// service within Window are collapsed into a single upstream call.
func NewCoalesceChecker(inner Checker, cfg CoalesceConfig, log *slog.Logger) Checker {
	if log == nil {
		log = slog.Default()
	}
	if cfg.Window <= 0 {
		cfg = DefaultCoalesceConfig()
	}
	return &coalesceChecker{inner: inner, cfg: cfg, log: log, flights: make(map[string]*coalesceFlight)}
}

func (c *coalesceChecker) Check(ctx context.Context, service string) Result {
	c.mu.Lock()
	if f, ok := c.flights[service]; ok {
		c.mu.Unlock()
		c.log.Debug("coalesce: waiting for in-flight check", "service", service)
		select {
		case <-f.doneCh:
			return f.result
		case <-ctx.Done():
			return Result{Status: StatusUnhealthy, Err: ctx.Err()}
		}
	}
	f := &coalesceFlight{doneCh: make(chan struct{})}
	c.flights[service] = f
	c.mu.Unlock()

	// Brief window to let stragglers join.
	select {
	case <-time.After(c.cfg.Window):
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.flights, service)
		c.mu.Unlock()
		close(f.doneCh)
		return Result{Status: StatusUnhealthy, Err: ctx.Err()}
	}

	f.result = c.inner.Check(ctx, service)
	c.mu.Lock()
	delete(c.flights, service)
	c.mu.Unlock()
	close(f.doneCh)
	return f.result
}
