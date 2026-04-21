package health

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

// DefaultGraceConfig returns a GraceConfig with sensible defaults.
func DefaultGraceConfig() GraceConfig {
	return GraceConfig{
		Window: 10 * time.Second,
		MinChecks: 3,
	}
}

// GraceConfig controls the startup grace period behaviour.
type GraceConfig struct {
	// Window is how long after the first check the grace period lasts.
	Window time.Duration
	// MinChecks is the minimum number of checks that must have been performed
	// before the grace period can be considered over regardless of Window.
	MinChecks int
	Logger    *slog.Logger
}

// graceChecker wraps an inner Checker and suppresses unhealthy results while
// the service is still within its startup grace period.
type graceChecker struct {
	cfg    GraceConfig
	inner  Checker
	count  atomic.Int64
	start  time.Time
	logged atomic.Bool
}

// NewGraceChecker returns a Checker that treats the service as healthy during
// the startup grace window, delegating to inner once the window has elapsed
// and the minimum number of checks has been reached.
func NewGraceChecker(inner Checker, cfg GraceConfig) Checker {
	if cfg.Window == 0 {
		cfg.Window = DefaultGraceConfig().Window
	}
	if cfg.MinChecks == 0 {
		cfg.MinChecks = DefaultGraceConfig().MinChecks
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &graceChecker{
		cfg:   cfg,
		inner: inner,
		start: time.Now(),
	}
}

func (g *graceChecker) Check(ctx context.Context, service string) Result {
	count := g.count.Add(1)
	result := g.inner.Check(ctx, service)

	inGrace := time.Since(g.start) < g.cfg.Window || count < int64(g.cfg.MinChecks)
	if inGrace && result.Status == StatusUnhealthy {
		if g.logged.CompareAndSwap(false, true) {
			g.cfg.Logger.Info("grace period active, suppressing unhealthy result",
				"service", service,
				"checks", count,
				"window", g.cfg.Window,
			)
		}
		return Result{Status: StatusHealthy, Message: "startup grace period active"}
	}

	return result
}
