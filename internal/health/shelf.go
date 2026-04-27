package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DefaultShelfConfig returns a ShelfConfig with sensible defaults.
func DefaultShelfConfig() ShelfConfig {
	return ShelfConfig{
		TTL:    30 * time.Second,
		Logger: slog.Default(),
	}
}

// ShelfConfig controls the behaviour of NewShelfChecker.
type ShelfConfig struct {
	// TTL is how long a successful result is held before being discarded.
	TTL time.Duration
	// Logger receives debug messages when a shelved result is served or evicted.
	Logger *slog.Logger
}

type shelfEntry struct {
	result Result
	expiresAt time.Time
}

// shelfChecker holds the last *healthy* result for a service and serves it
// while it remains within TTL. On unhealthy or error it always delegates.
type shelfChecker struct {
	inner  Checker
	cfg    ShelfConfig
	mu     sync.Mutex
	shelf  map[string]shelfEntry
}

// NewShelfChecker wraps inner so that a healthy result is cached for cfg.TTL
// and re-served on subsequent calls without hitting the upstream. Any
// unhealthy or error result bypasses the shelf and is returned directly.
func NewShelfChecker(inner Checker, cfg ShelfConfig) Checker {
	if cfg.TTL <= 0 {
		cfg.TTL = DefaultShelfConfig().TTL
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &shelfChecker{
		inner: inner,
		cfg:   cfg,
		shelf: make(map[string]shelfEntry),
	}
}

func (s *shelfChecker) Check(ctx context.Context, service string) Result {
	s.mu.Lock()
	if e, ok := s.shelf[service]; ok {
		if time.Now().Before(e.expiresAt) {
			s.mu.Unlock()
			s.cfg.Logger.Debug("shelf: serving cached healthy result", "service", service)
			return e.result
		}
		delete(s.shelf, service)
		s.cfg.Logger.Debug("shelf: evicted expired entry", "service", service)
	}
	s.mu.Unlock()

	res := s.inner.Check(ctx, service)
	if res.Status == StatusHealthy && res.Err == nil {
		s.mu.Lock()
		s.shelf[service] = shelfEntry{result: res, expiresAt: time.Now().Add(s.cfg.TTL)}
		s.mu.Unlock()
	}
	return res
}
