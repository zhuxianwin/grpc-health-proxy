package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DefaultCooldownConfig returns a CooldownConfig with sensible defaults.
func DefaultCooldownConfig() CooldownConfig {
	return CooldownConfig{
		Duration: 10 * time.Second,
	}
}

// CooldownConfig controls how long a service is held in cooldown after
// returning an unhealthy result.
type CooldownConfig struct {
	Duration time.Duration
}

type cooldownEntry struct {
	until time.Time
}

// CooldownChecker suppresses repeated checks for a service that recently
// returned unhealthy, returning the last known result until the cooldown
// period expires.
type CooldownChecker struct {
	inner  Checker
	cfg    CooldownConfig
	log    *slog.Logger
	mu     sync.Mutex
	cooled map[string]cooldownEntry
	last   map[string]Result
}

// NewCooldownChecker wraps inner with cooldown behaviour.
func NewCooldownChecker(inner Checker, cfg CooldownConfig, log *slog.Logger) *CooldownChecker {
	if cfg.Duration <= 0 {
		cfg = DefaultCooldownConfig()
	}
	if log == nil {
		log = slog.Default()
	}
	if cfg.Duration <= 0 {
		cfg = DefaultCooldownConfig()
	}
	return &CooldownChecker{
		inner:  inner,
		cfg:    cfg,
		log:    log,
		cooled: make(map[string]cooldownEntry),
		last:   make(map[string]Result),
	}
}

// Check delegates to the inner checker unless the service is in cooldown.
func (c *CooldownChecker) Check(ctx context.Context, service string) (Result, error) {
	c.mu.Lock()
	entry, cooling := c.cooled[service]
	if cooling && time.Now().Before(entry.until) {
		res := c.last[service]
		c.mu.Unlock()
		c.log.Debug("cooldown active, returning cached result", "service", service)
		return res, nil
	}
	delete(c.cooled, service)
	c.mu.Unlock()

	res, err := c.inner.Check(ctx, service)

	if err == nil && res.Status == StatusUnhealthy {
		c.mu.Lock()
		c.cooled[service] = cooldownEntry{until: time.Now().Add(c.cfg.Duration)}
		c.last[service] = res
		c.mu.Unlock()
		c.log.Debug("entering cooldown", "service", service, "duration", c.cfg.Duration)
	}
	return res, err
}
