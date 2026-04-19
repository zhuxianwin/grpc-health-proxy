package health

import (
	"context"
	"log/slog"
)

// DefaultClampConfig returns a ClampConfig with sensible defaults.
func DefaultClampConfig() ClampConfig {
	return ClampConfig{
		MinStatus: StatusHealthy,
		MaxStatus: StatusUnhealthy,
	}
}

// ClampConfig controls how ClampChecker constrains result status.
type ClampConfig struct {
	// MinStatus is the floor — results below this are raised to it.
	MinStatus Status
	// MaxStatus is the ceiling — results above this are lowered to it.
	MaxStatus Status
	Logger    *slog.Logger
}

// clampChecker wraps an inner Checker and clamps the returned Status.
type clampChecker struct {
	inner  Checker
	cfg    ClampConfig
	logger *slog.Logger
}

// NewClampChecker returns a Checker that clamps the status of inner results
// to the inclusive range [cfg.MinStatus, cfg.MaxStatus].
func NewClampChecker(inner Checker, cfg ClampConfig) Checker {
	if cfg.MinStatus == 0 && cfg.MaxStatus == 0 {
		cfg = DefaultClampConfig()
	}
	l := cfg.Logger
	if l == nil {
		l = slog.Default()
	}
	return &clampChecker{inner: inner, cfg: cfg, logger: l}
}

func (c *clampChecker) Check(ctx context.Context, service string) Result {
	res := c.inner.Check(ctx, service)
	orig := res.Status

	if res.Status < c.cfg.MinStatus {
		res.Status = c.cfg.MinStatus
	}
	if res.Status > c.cfg.MaxStatus {
		res.Status = c.cfg.MaxStatus
	}

	if res.Status != orig {
		c.logger.Debug("clamp: status adjusted",
			"service", service,
			"original", orig.String(),
			"clamped", res.Status.String(),
		)
	}
	return res
}
