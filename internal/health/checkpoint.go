package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// CheckpointConfig controls checkpoint behaviour.
type CheckpointConfig struct {
	// MinInterval is the minimum time between persisting a checkpoint.
	MinInterval time.Duration
}

// DefaultCheckpointConfig returns sensible defaults.
func DefaultCheckpointConfig() CheckpointConfig {
	return CheckpointConfig{MinInterval: 10 * time.Second}
}

// CheckpointEntry records the last known result for a service.
type CheckpointEntry struct {
	Service   string
	Result    Result
	RecordedAt time.Time
}

// CheckpointStore holds the most recent checkpointed result per service.
type CheckpointStore struct {
	mu      sync.RWMutex
	entries map[string]CheckpointEntry
}

// NewCheckpointStore creates an empty store.
func NewCheckpointStore() *CheckpointStore {
	return &CheckpointStore{entries: make(map[string]CheckpointEntry)}
}

// Record saves the result for the given service.
func (s *CheckpointStore) Record(service string, r Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[service] = CheckpointEntry{Service: service, Result: r, RecordedAt: time.Now()}
}

// Latest returns the most recent checkpoint for a service, if any.
func (s *CheckpointStore) Latest(service string) (CheckpointEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[service]
	return e, ok
}

// All returns a copy of all checkpointed entries.
func (s *CheckpointStore) All() []CheckpointEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]CheckpointEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}

// NewCheckpointChecker wraps inner and records every result into store.
func NewCheckpointChecker(inner Checker, store *CheckpointStore, cfg CheckpointConfig, log *slog.Logger) Checker {
	if cfg.MinInterval == 0 {
		cfg = DefaultCheckpointConfig()
	}
	if log == nil {
		log = slog.Default()
	}
	return &checkpointChecker{inner: inner, store: store, cfg: cfg, log: log}
}

type checkpointChecker struct {
	inner Checker
	store *CheckpointStore
	cfg   CheckpointConfig
	log   *slog.Logger
	mu    sync.Mutex
	last  map[string]time.Time
}

func (c *checkpointChecker) Check(ctx context.Context, service string) Result {
	r := c.inner.Check(ctx, service)
	if c.shouldRecord(service) {
		c.store.Record(service, r)
		c.log.Debug("checkpoint recorded", "service", service, "status", r.Status)
	}
	return r
}

func (c *checkpointChecker) shouldRecord(service string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.last == nil {
		c.last = make(map[string]time.Time)
	}
	if time.Since(c.last[service]) >= c.cfg.MinInterval {
		c.last[service] = time.Now()
		return true
	}
	return false
}
