package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// ThrottleConfig controls the minimum interval between upstream checks.
type ThrottleConfig struct {
	// MinInterval is the shortest time allowed between two checks for the same service.
	MinInterval time.Duration
}

// DefaultThrottleConfig returns a ThrottleConfig with sensible defaults.
func DefaultThrottleConfig() ThrottleConfig {
	return ThrottleConfig{MinInterval: 2 * time.Second}
}

type throttleEntry struct {
	last   time.Time
	result Result
}

// throttleChecker wraps a Checker and suppresses calls that arrive too soon
// after the previous one, returning the cached result instead.
type throttleChecker struct {
	next   Checker
	cfg    ThrottleConfig
	log    *slog.Logger
	mu     sync.Mutex
	state  map[string]throttleEntry
}

// NewThrottleChecker returns a Checker that enforces a minimum interval
// between upstream health calls per service name.
func NewThrottleChecker(next Checker, cfg ThrottleConfig, log *slog.Logger) Checker {
	if cfg.MinInterval <= 0 {
		cfg = DefaultThrottleConfig()
	}
	if log == nil {
		log = slog.Default()
	}
	return &throttleChecker{
		next:  next,
		cfg:   cfg,
		log:   log,
		state: make(map[string]throttleEntry),
	}
}

func (t *throttleChecker) Check(ctx context.Context, service string) Result {
	now := time.Now()

	t.mu.Lock()
	entry, ok := t.state[service]
	if ok && now.Sub(entry.last) < t.cfg.MinInterval {
		t.mu.Unlock()
		t.log.Debug("throttle: returning cached result", "service", service)
		return entry.result
	}
	t.mu.Unlock()

	res := t.next.Check(ctx, service)

	t.mu.Lock()
	t.state[service] = throttleEntry{last: time.Now(), result: res}
	t.mu.Unlock()

	return res
}
