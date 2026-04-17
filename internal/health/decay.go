package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DecayConfig controls how health results decay over time when a service
// repeatedly fails. After DecayAfter consecutive failures the checker
// introduces an exponentially growing cool-down before forwarding the
// next live probe, reducing thundering-herd pressure on a sick upstream.
type DecayConfig struct {
	// DecayAfter is the number of consecutive failures before decay kicks in.
	DecayAfter int
	// MaxDelay caps the cool-down duration.
	MaxDelay time.Duration
	// BaseDelay is the starting cool-down duration.
	BaseDelay time.Duration
}

// DefaultDecayConfig returns sensible defaults.
func DefaultDecayConfig() DecayConfig {
	return DecayConfig{
		DecayAfter: 3,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   30 * time.Second,
	}
}

type decayChecker struct {
	inner  Checker
	cfg    DecayConfig
	log    *slog.Logger
	mu     sync.Mutex
	fails  int
	nextAt time.Time
	last   Result
}

// NewDecayChecker wraps inner so that after cfg.DecayAfter consecutive
// failures subsequent calls are short-circuited and return the last
// cached failure until the computed cool-down elapses.
func NewDecayChecker(inner Checker, cfg DecayConfig, log *slog.Logger) Checker {
	if cfg.DecayAfter <= 0 {
		cfg = DefaultDecayConfig()
	}
	if log == nil {
		log = slog.Default()
	}
	return &decayChecker{inner: inner, cfg: cfg, log: log}
}

func (d *decayChecker) Check(ctx context.Context, service string) Result {
	d.mu.Lock()
	if d.fails >= d.cfg.DecayAfter && time.Now().Before(d.nextAt) {
		cached := d.last
		d.mu.Unlock()
		d.log.Debug("decay: returning cached failure", "service", service, "next_probe", d.nextAt)
		return cached
	}
	d.mu.Unlock()

	res := d.inner.Check(ctx, service)

	d.mu.Lock()
	defer d.mu.Unlock()
	if res.Err != nil || res.Status == StatusUnhealthy {
		d.fails++
		delay := d.cfg.BaseDelay * (1 << min(d.fails-d.cfg.DecayAfter, 10))
		if delay > d.cfg.MaxDelay {
			delay = d.cfg.MaxDelay
		}
		if d.fails >= d.cfg.DecayAfter {
			d.nextAt = time.Now().Add(delay)
		}
	} else {
		d.fails = 0
		d.nextAt = time.Time{}
	}
	d.last = res
	return res
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
