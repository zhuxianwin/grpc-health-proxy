package health

import (
	"log/slog"
	"sync"
	"time"
)

// SnapshotConfig controls snapshot behaviour.
type SnapshotConfig struct {
	// Interval between automatic snapshots. Defaults to 30s.
	Interval time.Duration
}

// DefaultSnapshotConfig returns sensible defaults.
func DefaultSnapshotConfig() SnapshotConfig {
	return SnapshotConfig{Interval: 30 * time.Second}
}

// SnapshotEntry holds a point-in-time result for a service.
type SnapshotEntry struct {
	Service   string
	Result    Result
	RecordedAt time.Time
}

// SnapshotStore records the most recent result per service.
type SnapshotStore struct {
	mu      sync.RWMutex
	entries map[string]SnapshotEntry
	log     *slog.Logger
}

// NewSnapshotStore creates an empty store.
func NewSnapshotStore(log *slog.Logger) *SnapshotStore {
	if log == nil {
		log = slog.Default()
	}
	return &SnapshotStore{
		entries: make(map[string]SnapshotEntry),
		log:     log,
	}
}

// Record saves a result for the given service.
func (s *SnapshotStore) Record(service string, r Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[service] = SnapshotEntry{
		Service:    service,
		Result:     r,
		RecordedAt: time.Now(),
	}
	s.log.Debug("snapshot recorded", "service", service, "status", r.Status)
}

// Latest returns the most recent snapshot for a service, if any.
func (s *SnapshotStore) Latest(service string) (SnapshotEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[service]
	return e, ok
}

// All returns a copy of every recorded snapshot.
func (s *SnapshotStore) All() []SnapshotEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SnapshotEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
