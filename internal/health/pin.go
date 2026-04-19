package health

import (
	"context"
	"log/slog"
	"sync"
)

// PinConfig controls the pin checker behaviour.
type PinConfig struct {
	// Logger is optional; defaults to slog.Default().
	Logger *slog.Logger
}

// DefaultPinConfig returns a PinConfig with sensible defaults.
func DefaultPinConfig() PinConfig {
	return PinConfig{
		Logger: slog.Default(),
	}
}

// PinChecker wraps an inner Checker and allows the result to be pinned
// (overridden) to a fixed Result for testing or maintenance windows.
type PinChecker struct {
	inner  Checker
	cfg    PinConfig
	mu     sync.RWMutex
	pinned *Result
}

// NewPinChecker creates a PinChecker wrapping inner.
func NewPinChecker(inner Checker, cfg PinConfig) *PinChecker {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &PinChecker{inner: inner, cfg: cfg}
}

// Pin fixes the result returned for all subsequent Check calls.
func (p *PinChecker) Pin(r Result) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pinned = &r
	p.cfg.Logger.Info("health check pinned", "status", r.Status)
}

// Unpin removes any previously pinned result.
func (p *PinChecker) Unpin() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pinned = nil
	p.cfg.Logger.Info("health check unpinned")
}

// Pinned returns the currently pinned result, or nil if not pinned.
func (p *PinChecker) Pinned() *Result {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.pinned == nil {
		return nil
	}
	copy := *p.pinned
	return &copy
}

// Check returns the pinned result when set, otherwise delegates to inner.
func (p *PinChecker) Check(ctx context.Context, service string) (Result, error) {
	p.mu.RLock()
	pin := p.pinned
	p.mu.RUnlock()
	if pin != nil {
		return *pin, nil
	}
	return p.inner.Check(ctx, service)
}
