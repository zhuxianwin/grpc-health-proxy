package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DefaultFenceConfig returns a FenceConfig with sensible defaults.
func DefaultFenceConfig() FenceConfig {
	return FenceConfig{
		OpenDuration: 10 * time.Second,
		Logger:       slog.Default(),
	}
}

// FenceConfig controls the behaviour of the FenceChecker.
type FenceConfig struct {
	// OpenDuration is how long the fence remains open (blocking checks)
	// after being raised. Zero uses the default.
	OpenDuration time.Duration

	Logger *slog.Logger
}

// FenceChecker blocks health checks while a fence is raised, returning an
// unhealthy result immediately. Once the fence is lowered (or its duration
// expires) checks are delegated to the inner Checker as normal.
type FenceChecker struct {
	inner  Checker
	cfg    FenceConfig
	mu     sync.Mutex
	rainedAt time.Time
	raised  bool
}

// NewFenceChecker wraps inner with a manually-controlled fence. If cfg is
// zero-valued, DefaultFenceConfig is used.
func NewFenceChecker(inner Checker, cfg FenceConfig) *FenceChecker {
	if cfg.OpenDuration == 0 {
		cfg = DefaultFenceConfig()
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &FenceChecker{inner: inner, cfg: cfg}
}

// Raise lifts the fence, causing subsequent Check calls to return unhealthy
// until Lower is called or OpenDuration elapses.
func (f *FenceChecker) Raise() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.raised = true
	f.rainedAt = time.Now()
	f.cfg.Logger.Info("fence raised")
}

// Lower drops the fence, restoring normal delegation.
func (f *FenceChecker) Lower() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.raised = false
	f.cfg.Logger.Info("fence lowered")
}

// IsRaised reports whether the fence is currently active.
func (f *FenceChecker) IsRaised() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.raised && time.Since(f.rainedAt) >= f.cfg.OpenDuration {
		f.raised = false
		f.cfg.Logger.Info("fence expired, auto-lowered")
	}
	return f.raised
}

// Check implements Checker. Returns unhealthy while the fence is raised.
func (f *FenceChecker) Check(ctx context.Context, service string) Result {
	if f.IsRaised() {
		return Result{Status: StatusUnhealthy, Err: ErrFenceRaised}
	}
	return f.inner.Check(ctx, service)
}
