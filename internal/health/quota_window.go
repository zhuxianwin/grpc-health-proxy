package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DefaultQuotaWindowConfig returns a QuotaWindowConfig with sensible defaults.
func DefaultQuotaWindowConfig() QuotaWindowConfig {
	return QuotaWindowConfig{
		Window:  time.Minute,
		MaxCalls: 60,
	}
}

// QuotaWindowConfig configures the sliding-window quota checker.
type QuotaWindowConfig struct {
	// Window is the rolling duration over which calls are counted.
	Window time.Duration
	// MaxCalls is the maximum number of health checks allowed per window.
	MaxCalls int
	// Logger is optional; defaults to slog.Default().
	Logger *slog.Logger
}

type quotaWindowChecker struct {
	cfg    QuotaWindowConfig
	inner  Checker
	mu     sync.Mutex
	log    *slog.Logger
	// timestamps of recent calls per service
	times  map[string][]time.Time
}

// NewQuotaWindowChecker wraps inner and enforces a sliding-window call quota.
// When the quota is exceeded the last known result is returned without calling
// the inner checker.
func NewQuotaWindowChecker(inner Checker, cfg QuotaWindowConfig) Checker {
	if cfg.Window <= 0 {
		cfg.Window = DefaultQuotaWindowConfig().Window
	}
	if cfg.MaxCalls <= 0 {
		cfg.MaxCalls = DefaultQuotaWindowConfig().MaxCalls
	}
	log := cfg.Logger
	if log == nil {
		log = slog.Default()
	}
	return &quotaWindowChecker{
		cfg:   cfg,
		inner: inner,
		log:   log,
		times: make(map[string][]time.Time),
	}
}

func (q *quotaWindowChecker) Check(ctx context.Context, service string) Result {
	now := time.Now()
	cutoff := now.Add(-q.cfg.Window)

	q.mu.Lock()
	ts := q.times[service]
	// evict entries outside the window
	valid := ts[:0]
	for _, t := range ts {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	exceeded := len(valid) >= q.cfg.MaxCalls
	if !exceeded {
		valid = append(valid, now)
	}
	q.times[service] = valid
	q.mu.Unlock()

	if exceeded {
		q.log.Warn("quota_window: limit exceeded, skipping check",
			"service", service,
			"max_calls", q.cfg.MaxCalls,
			"window", q.cfg.Window,
		)
		return Result{Status: StatusUnknown, Err: ErrQuotaExceeded}
	}
	return q.inner.Check(ctx, service)
}
