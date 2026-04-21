package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DefaultBurstConfig returns a BurstConfig with sensible defaults.
func DefaultBurstConfig() BurstConfig {
	return BurstConfig{
		MaxBurst:  5,
		Window:    10 * time.Second,
		Logger:    slog.Default(),
	}
}

// BurstConfig controls burst-detection behaviour.
type BurstConfig struct {
	// MaxBurst is the maximum number of checks allowed within Window before
	// the checker starts returning Unhealthy.
	MaxBurst int
	// Window is the rolling time window over which bursts are counted.
	Window time.Duration
	// Logger receives debug messages when burst is exceeded.
	Logger *slog.Logger
}

type burstChecker struct {
	inner  Checker
	cfg    BurstConfig
	mu     sync.Mutex
	times  []time.Time
}

// NewBurstChecker wraps inner and returns Unhealthy when the number of
// health-check calls within cfg.Window exceeds cfg.MaxBurst.
func NewBurstChecker(inner Checker, cfg BurstConfig) Checker {
	if cfg.MaxBurst <= 0 {
		cfg.MaxBurst = DefaultBurstConfig().MaxBurst
	}
	if cfg.Window <= 0 {
		cfg.Window = DefaultBurstConfig().Window
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &burstChecker{inner: inner, cfg: cfg}
}

func (b *burstChecker) Check(ctx context.Context, service string) Result {
	now := time.Now()
	b.mu.Lock()
	cutoff := now.Add(-b.cfg.Window)
	filtered := b.times[:0]
	for _, t := range b.times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	filtered = append(filtered, now)
	b.times = filtered
	count := len(b.times)
	b.mu.Unlock()

	if count > b.cfg.MaxBurst {
		b.cfg.Logger.Debug("burst limit exceeded",
			"service", service,
			"count", count,
			"max_burst", b.cfg.MaxBurst,
		)
		return Result{
			Status:  StatusUnhealthy,
			Message: "burst limit exceeded",
		}
	}
	return b.inner.Check(ctx, service)
}
