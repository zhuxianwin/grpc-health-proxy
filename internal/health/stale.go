package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// StaleConfig controls stale-while-revalidate behaviour.
type StaleConfig struct {
	// StaleFor is how long a stale result may be served after expiry.
	StaleFor time.Duration
}

// DefaultStaleConfig returns a StaleConfig with sensible defaults.
func DefaultStaleConfig() StaleConfig {
	return StaleConfig{StaleFor: 10 * time.Second}
}

type staleEntry struct {
	result  Result
	cachedAt time.Time
}

// StaleChecker wraps a Checker and serves the last known result while a
// background revalidation is in flight, up to StaleFor after expiry.
type StaleChecker struct {
	inner  Checker
	cfg    StaleConfig
	log    *slog.Logger

	mu      sync.Mutex
	entries map[string]*staleEntry
	flying  map[string]bool
}

// NewStaleChecker returns a StaleChecker wrapping inner.
func NewStaleChecker(inner Checker, cfg StaleConfig, log *slog.Logger) *StaleChecker {
	if cfg.StaleFor == 0 {
		cfg = DefaultStaleConfig()
	}
	if log == nil {
		log = slog.Default()
	}
	return &StaleChecker{
		inner:   inner,
		cfg:     cfg,
		log:     log,
		entries: make(map[string]*staleEntry),
		flying:  make(map[string]bool),
	}
}

// Check returns a cached result if one exists within the stale window,
// triggering a background refresh. Otherwise it calls inner directly.
func (s *StaleChecker) Check(ctx context.Context, service string) Result {
	s.mu.Lock()
	e, ok := s.entries[service]
	if ok && time.Since(e.cachedAt) <= s.cfg.StaleFor {
		if !s.flying[service] {
			s.flying[service] = true
			go s.revalidate(service)
		}
		s.mu.Unlock()
		s.log.Debug("stale: serving cached result", "service", service)
		return e.result
	}
	s.mu.Unlock()

	r := s.inner.Check(ctx, service)
	s.mu.Lock()
	s.entries[service] = &staleEntry{result: r, cachedAt: time.Now()}
	s.mu.Unlock()
	return r
}

func (s *StaleChecker) revalidate(service string) {
	r := s.inner.Check(context.Background(), service)
	s.mu.Lock()
	s.entries[service] = &staleEntry{result: r, cachedAt: time.Now()}
	delete(s.flying, service)
	s.mu.Unlock()
	s.log.Debug("stale: revalidated", "service", service, "healthy", r.Healthy)
}
