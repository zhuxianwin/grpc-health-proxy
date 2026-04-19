package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// MuteConfig controls how long a checker is silenced after being muted.
type MuteConfig struct {
	// Duration is how long the checker stays muted. Defaults to 30s.
	Duration time.Duration
}

// DefaultMuteConfig returns sensible defaults.
func DefaultMuteConfig() MuteConfig {
	return MuteConfig{Duration: 30 * time.Second}
}

// MuteChecker wraps a Checker and allows it to be temporarily muted,
// returning a healthy result while muted.
type MuteChecker struct {
	inner  Checker
	cfg    MuteConfig
	log    *slog.Logger
	mu     sync.Mutex
	mutedUntil time.Time
}

// NewMuteChecker creates a MuteChecker wrapping inner.
func NewMuteChecker(inner Checker, cfg MuteConfig, log *slog.Logger) *MuteChecker {
	if log == nil {
		log = slog.Default()
	}
	if cfg.Duration == 0 {
		cfg = DefaultMuteConfig()
	}
	return &MuteChecker{inner: inner, cfg: cfg, log: log}
}

// Mute silences the checker for the configured duration.
func (m *MuteChecker) Mute() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mutedUntil = time.Now().Add(m.cfg.Duration)
	m.log.Info("checker muted", "until", m.mutedUntil)
}

// Unmute cancels an active mute immediately.
func (m *MuteChecker) Unmute() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mutedUntil = time.Time{}
}

// IsMuted reports whether the checker is currently muted.
func (m *MuteChecker) IsMuted() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return time.Now().Before(m.mutedUntil)
}

// Check delegates to inner unless muted, in which case it returns healthy.
func (m *MuteChecker) Check(ctx context.Context, service string) Result {
	if m.IsMuted() {
		m.log.Debug("check suppressed by mute", "service", service)
		return Result{Status: StatusHealthy, Message: "muted"}
	}
	return m.inner.Check(ctx, service)
}
