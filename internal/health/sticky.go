package health

import (
	"context"
	"sync"
	"time"
)

// DefaultStickyConfig returns a StickyConfig with sensible defaults.
func DefaultStickyConfig() StickyConfig {
	return StickyConfig{
		UnhealthyTTL: 30 * time.Second,
	}
}

// StickyConfig controls how long an unhealthy result is held.
type StickyConfig struct {
	// UnhealthyTTL is how long an unhealthy result is "stuck" before re-checking.
	UnhealthyTTL time.Duration
}

type stickyEntry struct {
	result Result
	stuckUntil time.Time
}

type stickyChecker struct {
	inner  Checker
	cfg    StickyConfig
	mu     sync.Mutex
	entries map[string]stickyEntry
}

// NewStickyChecker wraps a Checker so that once a service becomes unhealthy,
// that result is held for UnhealthyTTL before the inner checker is consulted again.
func NewStickyChecker(inner Checker, cfg StickyConfig) Checker {
	if cfg.UnhealthyTTL <= 0 {
		cfg = DefaultStickyConfig()
	}
	return &stickyChecker{
		inner:   inner,
		cfg:     cfg,
		entries: make(map[string]stickyEntry),
	}
}

func (s *stickyChecker) Check(ctx context.Context, service string) (Result, error) {
	s.mu.Lock()
	entry, ok := s.entries[service]
	s.mu.Unlock()

	if ok && entry.result.Status == StatusUnhealthy && time.Now().Before(entry.stuckUntil) {
		return entry.result, nil
	}

	res, err := s.inner.Check(ctx, service)
	if err != nil {
		return res, err
	}

	if res.Status == StatusUnhealthy {
		s.mu.Lock()
		s.entries[service] = stickyEntry{
			result:     res,
			stuckUntil: time.Now().Add(s.cfg.UnhealthyTTL),
		}
		s.mu.Unlock()
	} else {
		s.mu.Lock()
		delete(s.entries, service)
		s.mu.Unlock()
	}

	return res, nil
}
