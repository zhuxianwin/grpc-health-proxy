package health

import (
	"context"
	"log/slog"
	"sync/atomic"
)

// DefaultSkipConfig returns a SkipConfig with safe defaults.
func DefaultSkipConfig() SkipConfig {
	return SkipConfig{
		Logger: slog.Default(),
	}
}

// SkipConfig controls the behaviour of a SkipChecker.
type SkipConfig struct {
	// Logger is used to emit a debug line when a check is skipped.
	Logger *slog.Logger
}

// SkipChecker wraps a Checker and allows callers to toggle whether checks are
// forwarded to the inner checker or silently skipped (returning the last known
// result, or Healthy when no result is cached).
type SkipChecker struct {
	inner  Checker
	cfg    SkipConfig
	skip   atomic.Bool
	last   atomic.Pointer[Result]
}

// NewSkipChecker creates a SkipChecker wrapping inner.
func NewSkipChecker(inner Checker, cfg SkipConfig) *SkipChecker {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &SkipChecker{inner: inner, cfg: cfg}
}

// Skip enables skipping so that subsequent Check calls bypass the inner
// checker and return the last known result.
func (s *SkipChecker) Skip() { s.skip.Store(true) }

// Resume disables skipping and restores normal delegation.
func (s *SkipChecker) Resume() { s.skip.Store(false) }

// Skipping reports whether the checker is currently in skip mode.
func (s *SkipChecker) Skipping() bool { return s.skip.Load() }

// Check implements Checker. When skipping is active the last known Result is
// returned without calling the inner checker. If no previous result exists a
// Healthy result is synthesised.
func (s *SkipChecker) Check(ctx context.Context, service string) Result {
	if s.skip.Load() {
		s.cfg.Logger.Debug("health check skipped", "service", service)
		if prev := s.last.Load(); prev != nil {
			return *prev
		}
		return Result{Status: StatusHealthy}
	}
	r := s.inner.Check(ctx, service)
	s.last.Store(&r)
	return r
}
