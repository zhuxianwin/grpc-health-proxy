package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DefaultDebounceConfig returns a DebounceConfig with sensible defaults.
func DefaultDebounceConfig() DebounceConfig {
	return DebounceConfig{
		Window: 5 * time.Second,
	}
}

// DebounceConfig controls debounce behaviour.
type DebounceConfig struct {
	// Window is the minimum time between consecutive upstream checks.
	Window time.Duration
}

type debounceChecker struct {
	inner  Checker
	cfg    DebounceConfig
	log    *slog.Logger

	mu       sync.Mutex
	lastTime map[string]time.Time
	lastRes  map[string]Result
}

// NewDebounceChecker wraps inner so that repeated calls within Window reuse
// the previous result instead of hitting the upstream service again.
func NewDebounceChecker(inner Checker, cfg DebounceConfig, log *slog.Logger) Checker {
	if cfg.Window <= 0 {
		cfg = DefaultDebounceConfig()
	}
	if log == nil {
		log = slog.Default()
	}
	return &debounceChecker{
		inner:    inner,
		cfg:      cfg,
		log:      log,
		lastTime: make(map[string]time.Time),
		lastRes:  make(map[string]Result),
	}
}

func (d *debounceChecker) Check(ctx context.Context, service string) (Result, error) {
	d.mu.Lock()
	if t, ok := d.lastTime[service]; ok && time.Since(t) < d.cfg.Window {
		res := d.lastRes[service]
		d.mu.Unlock()
		d.log.Debug("debounce: returning cached result", "service", service)
		return res, nil
	}
	d.mu.Unlock()

	res, err := d.inner.Check(ctx, service)

	d.mu.Lock()
	d.lastTime[service] = time.Now()
	d.lastRes[service] = res
	d.mu.Unlock()

	return res, err
}
