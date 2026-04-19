package health

import (
	"context"
	"log/slog"
)

// DefaultCapConfig returns a CapConfig with sensible defaults.
func DefaultCapConfig() CapConfig {
	return CapConfig{
		Max: StatusHealthy,
	}
}

// CapConfig controls the maximum status a CapChecker will return.
type CapConfig struct {
	// Max is the highest Status that will be returned. Results with a higher
	// numeric status are clamped to Max.
	Max    Status
	Logger *slog.Logger
}

// capChecker wraps an inner Checker and ensures the returned Status never
// exceeds a configured ceiling.
type capChecker struct {
	inner  Checker
	cfg    CapConfig
	logger *slog.Logger
}

// NewCapChecker returns a Checker that caps the status returned by inner at
// cfg.Max. Any status numerically greater than Max is replaced with Max.
// A nil logger falls back to the default slog logger.
func NewCapChecker(inner Checker, cfg CapConfig) Checker {
	if cfg.Max == 0 {
		cfg = DefaultCapConfig()
	}
	l := cfg.Logger
	if l == nil {
		l = slog.Default()
	}
	return &capChecker{inner: inner, cfg: cfg, logger: l}
}

func (c *capChecker) Check(ctx context.Context, service string) Result {
	res := c.inner.Check(ctx, service)
	if res.Status > c.cfg.Max {
		c.logger.Debug("cap: clamping status",
			"service", service,
			"original", res.Status,
			"cap", c.cfg.Max,
		)
		res.Status = c.cfg.Max
	}
	return res
}
