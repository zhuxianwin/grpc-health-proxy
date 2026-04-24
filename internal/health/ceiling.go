package health

import (
	"context"
	"log/slog"
	"sync"
)

// DefaultCeilingConfig returns a CeilingConfig with sensible defaults.
func DefaultCeilingConfig() CeilingConfig {
	return CeilingConfig{
		MaxUnhealthy: 1,
	}
}

// CeilingConfig controls the ceiling checker behaviour.
type CeilingConfig struct {
	// MaxUnhealthy is the maximum number of consecutive unhealthy results
	// before the checker begins returning a capped unhealthy status.
	MaxUnhealthy int
	// Logger is an optional structured logger.
	Logger *slog.Logger
}

// ceilingChecker wraps an inner Checker and caps the number of consecutive
// unhealthy results that are forwarded to callers. Once the ceiling is
// reached every subsequent check returns the last unhealthy result without
// delegating to the inner checker again until a healthy result is observed.
type ceilingChecker struct {
	inner  Checker
	cfg    CeilingConfig
	mu     sync.Mutex
	count  int
	last   *Result
	log    *slog.Logger
}

// NewCeilingChecker wraps inner with a ceiling that suppresses repeated
// unhealthy checks once cfg.MaxUnhealthy consecutive failures have been seen.
func NewCeilingChecker(inner Checker, cfg CeilingConfig) Checker {
	if cfg.MaxUnhealthy <= 0 {
		cfg.MaxUnhealthy = DefaultCeilingConfig().MaxUnhealthy
	}
	l := cfg.Logger
	if l == nil {
		l = slog.Default()
	}
	return &ceilingChecker{inner: inner, cfg: cfg, log: l}
}

func (c *ceilingChecker) Check(ctx context.Context, service string) Result {
	c.mu.Lock()
	if c.count >= c.cfg.MaxUnhealthy && c.last != nil {
		r := *c.last
		c.mu.Unlock()
		c.log.Debug("ceiling: suppressing check, returning cached unhealthy result",
			"service", service, "count", c.count)
		return r
	}
	c.mu.Unlock()

	r := c.inner.Check(ctx, service)

	c.mu.Lock()
	defer c.mu.Unlock()
	if r.Status == StatusHealthy {
		c.count = 0
		c.last = nil
	} else {
		c.count++
		copy := r
		c.last = &copy
	}
	return r
}
