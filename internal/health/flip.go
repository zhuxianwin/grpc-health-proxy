package health

import (
	"context"
	"log/slog"
	"sync"
)

// FlipConfig controls the behaviour of NewFlipChecker.
type FlipConfig struct {
	// Logger is optional; a default logger is used when nil.
	Logger *slog.Logger
}

// DefaultFlipConfig returns a FlipConfig with sensible defaults.
func DefaultFlipConfig() FlipConfig {
	return FlipConfig{
		Logger: slog.Default(),
	}
}

// flipChecker inverts the Status of the inner checker's result.
// Healthy becomes Unhealthy and vice-versa; errors are passed through.
type flipChecker struct {
	inner  Checker
	cfg    FlipConfig
	mu     sync.Mutex
	active bool
}

// NewFlipChecker wraps inner so that its status is inverted whenever
// flipping is active. Call Toggle to enable/disable inversion at runtime.
func NewFlipChecker(inner Checker, cfg FlipConfig) *flipChecker {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &flipChecker{inner: inner, cfg: cfg}
}

// Toggle switches the inversion on or off and returns the new state.
func (f *flipChecker) Toggle() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.active = !f.active
	f.cfg.Logger.Info("flip toggled", "active", f.active)
	return f.active
}

// Active reports whether inversion is currently enabled.
func (f *flipChecker) Active() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.active
}

func (f *flipChecker) Check(ctx context.Context, service string) Result {
	res := f.inner.Check(ctx, service)
	if res.Err != nil {
		return res
	}
	f.mu.Lock()
	active := f.active
	f.mu.Unlock()
	if !active {
		return res
	}
	if res.Status == StatusHealthy {
		res.Status = StatusUnhealthy
	} else {
		res.Status = StatusHealthy
	}
	return res
}
