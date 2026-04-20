package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DefaultLatchConfig returns a LatchConfig with sensible defaults.
func DefaultLatchConfig() LatchConfig {
	return LatchConfig{
		UnhealthyThreshold: 3,
		ResetAfter:         30 * time.Second,
	}
}

// LatchConfig controls when the latch trips and auto-resets.
type LatchConfig struct {
	// UnhealthyThreshold is the number of consecutive unhealthy results
	// required before the latch trips and begins returning unhealthy
	// without delegating.
	UnhealthyThreshold int

	// ResetAfter is how long after tripping before the latch auto-resets
	// and delegates again. Zero disables auto-reset.
	ResetAfter time.Duration
}

// latchChecker holds consecutive-failure state and trips after a threshold.
type latchChecker struct {
	inner  Checker
	cfg    LatchConfig
	log    *slog.Logger
	mu     sync.Mutex
	count  int
	tripped bool
	trippedAt time.Time
}

// NewLatchChecker wraps inner and trips after cfg.UnhealthyThreshold
// consecutive unhealthy results, returning unhealthy without delegating
// until the latch is reset.
func NewLatchChecker(inner Checker, cfg LatchConfig, log *slog.Logger) Checker {
	if cfg.UnhealthyThreshold <= 0 {
		cfg = DefaultLatchConfig()
	}
	if log == nil {
		log = slog.Default()
	}
	return &latchChecker{inner: inner, cfg: cfg, log: log}
}

func (l *latchChecker) Check(ctx context.Context, service string) Result {
	l.mu.Lock()
	if l.tripped {
		if l.cfg.ResetAfter > 0 && time.Since(l.trippedAt) >= l.cfg.ResetAfter {
			l.log.Info("latch auto-reset", "service", service)
			l.tripped = false
			l.count = 0
		} else {
			l.mu.Unlock()
			return Result{Status: StatusUnhealthy, Err: ErrLatchTripped}
		}
	}
	l.mu.Unlock()

	res := l.inner.Check(ctx, service)

	l.mu.Lock()
	defer l.mu.Unlock()
	if res.Status == StatusUnhealthy || res.Err != nil {
		l.count++
		if l.count >= l.cfg.UnhealthyThreshold && !l.tripped {
			l.tripped = true
			l.trippedAt = time.Now()
			l.log.Warn("latch tripped", "service", service, "threshold", l.cfg.UnhealthyThreshold)
		}
	} else {
		l.count = 0
	}
	return res
}

// Reset clears the latch and consecutive-failure counter.
func (l *latchChecker) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tripped = false
	l.count = 0
}

// Tripped returns true if the latch is currently tripped.
func (l *latchChecker) Tripped() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.tripped
}
