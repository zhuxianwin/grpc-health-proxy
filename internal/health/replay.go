package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DefaultReplayConfig returns a ReplayConfig with sensible defaults.
func DefaultReplayConfig() ReplayConfig {
	return ReplayConfig{
		Window:   30 * time.Second,
		Capacity: 100,
	}
}

// ReplayConfig controls the replay checker behaviour.
type ReplayConfig struct {
	// Window is how far back recorded results are retained.
	Window time.Duration
	// Capacity is the maximum number of results stored per service.
	Capacity int
	Logger   *slog.Logger
}

type replayEntry struct {
	result Result
	at     time.Time
}

// ReplayChecker records every result from the inner Checker and can replay
// the most-recent N results within a sliding time window.
type ReplayChecker struct {
	cfg   ReplayConfig
	inner Checker
	mu    sync.Mutex
	log   map[string][]replayEntry
}

// NewReplayChecker wraps inner with result recording.
func NewReplayChecker(inner Checker, cfg ReplayConfig) *ReplayChecker {
	if cfg.Window == 0 {
		cfg.Window = DefaultReplayConfig().Window
	}
	if cfg.Capacity == 0 {
		cfg.Capacity = DefaultReplayConfig().Capacity
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &ReplayChecker{cfg: cfg, inner: inner, log: make(map[string][]replayEntry)}
}

// Check delegates to the inner Checker and records the result.
func (r *ReplayChecker) Check(ctx context.Context, service string) (Result, error) {
	res, err := r.inner.Check(ctx, service)
	r.record(service, res)
	return res, err
}

func (r *ReplayChecker) record(service string, res Result) {
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()
	entries := r.evict(r.log[service], now)
	if len(entries) >= r.cfg.Capacity {
		entries = entries[1:]
	}
	r.log[service] = append(entries, replayEntry{result: res, at: now})
}

func (r *ReplayChecker) evict(entries []replayEntry, now time.Time) []replayEntry {
	cutoff := now.Add(-r.cfg.Window)
	i := 0
	for i < len(entries) && entries[i].at.Before(cutoff) {
		i++
	}
	return entries[i:]
}

// History returns recorded results for service within the configured window.
func (r *ReplayChecker) History(service string) []Result {
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()
	entries := r.evict(r.log[service], now)
	r.log[service] = entries
	out := make([]Result, len(entries))
	for i, e := range entries {
		out[i] = e.result
	}
	return out
}
