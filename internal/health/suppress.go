package health

import (
	"context"
	"log/slog"
	"sync"
)

// SuppressConfig controls which status values are suppressed (replaced with Healthy).
type SuppressConfig struct {
	// Statuses lists the Status values to suppress.
	Statuses []Status
	// Logger is optional; defaults to slog.Default().
	Logger *slog.Logger
}

// DefaultSuppressConfig returns a SuppressConfig that suppresses nothing.
func DefaultSuppressConfig() SuppressConfig {
	return SuppressConfig{
		Statuses: nil,
		Logger:   slog.Default(),
	}
}

// SuppressChecker wraps an inner Checker and replaces configured statuses with Healthy.
type SuppressChecker struct {
	inner  Checker
	cfg    SuppressConfig
	set    map[Status]struct{}
	mu     sync.RWMutex
}

// NewSuppressChecker creates a SuppressChecker wrapping inner.
func NewSuppressChecker(inner Checker, cfg SuppressConfig) *SuppressChecker {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	set := make(map[Status]struct{}, len(cfg.Statuses))
	for _, s := range cfg.Statuses {
		set[s] = struct{}{}
	}
	return &SuppressChecker{inner: inner, cfg: cfg, set: set}
}

// Check delegates to the inner checker and suppresses configured statuses.
func (s *SuppressChecker) Check(ctx context.Context, service string) Result {
	res := s.inner.Check(ctx, service)
	s.mu.RLock()
	_, suppressed := s.set[res.Status]
	s.mu.RUnlock()
	if suppressed {
		s.cfg.Logger.Debug("suppressing status",
			"service", service,
			"original_status", res.Status.String())
		res.Status = StatusHealthy
		res.Err = nil
	}
	return res
}

// AddStatus dynamically adds a status to the suppression set.
func (s *SuppressChecker) AddStatus(st Status) {
	s.mu.Lock()
	s.set[st] = struct{}{}
	s.mu.Unlock()
}

// RemoveStatus dynamically removes a status from the suppression set.
func (s *SuppressChecker) RemoveStatus(st Status) {
	s.mu.Lock()
	delete(s.set, st)
	s.mu.Unlock()
}
