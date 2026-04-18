package health

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

// DrainConfig controls drain behaviour.
type DrainConfig struct {
	// GracePeriod is how long to report unhealthy before shutdown.
	GracePeriod time.Duration
}

// DefaultDrainConfig returns sensible defaults.
func DefaultDrainConfig() DrainConfig {
	return DrainConfig{GracePeriod: 10 * time.Second}
}

// DrainChecker wraps a Checker and reports unhealthy once draining has begun,
// allowing load balancers to remove the pod before the process exits.
type DrainChecker struct {
	inner   Checker
	config  DrainConfig
	draining atomic.Bool
	log     *slog.Logger
}

// NewDrainChecker creates a DrainChecker. If cfg is zero the defaults are used.
func NewDrainChecker(inner Checker, cfg DrainConfig, log *slog.Logger) *DrainChecker {
	if cfg.GracePeriod == 0 {
		cfg = DefaultDrainConfig()
	}
	if log == nil {
		log = slog.Default()
	}
	return &DrainChecker{inner: inner, config: cfg, log: log}
}

// Check returns unhealthy immediately when draining, otherwise delegates.
func (d *DrainChecker) Check(ctx context.Context, service string) Result {
	if d.draining.Load() {
		return Result{Status: StatusUnhealthy}
	}
	return d.inner.Check(ctx, service)
}

// Drain begins the drain sequence. It marks the checker as draining and blocks
// for the configured grace period so callers can wait before shutting down.
func (d *DrainChecker) Drain(ctx context.Context) {
	d.draining.Store(true)
	d.log.Info("drain started", "grace_period", d.config.GracePeriod)
	select {
	case <-time.After(d.config.GracePeriod):
		d.log.Info("drain grace period elapsed")
	case <-ctx.Done():
		d.log.Warn("drain interrupted by context", "err", ctx.Err())
	}
}

// Draining reports whether the checker is currently draining.
func (d *DrainChecker) Draining() bool { return d.draining.Load() }
