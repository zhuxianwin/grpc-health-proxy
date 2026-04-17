package health

import (
	"context"
	"log/slog"
	"time"
)

// ObserveConfig holds configuration for the ObserveChecker.
type ObserveConfig struct {
	// OnResult is called after every check with the service name, result, and elapsed duration.
	OnResult func(service string, result Result, elapsed time.Duration)
}

// DefaultObserveConfig returns an ObserveConfig with a no-op callback.
func DefaultObserveConfig() ObserveConfig {
	return ObserveConfig{
		OnResult: func(_ string, _ Result, _ time.Duration) {},
	}
}

type observeChecker struct {
	inner  Checker
	cfg    ObserveConfig
	logger *slog.Logger
}

// NewObserveChecker wraps inner and calls cfg.OnResult after each check.
// If cfg.OnResult is nil, DefaultObserveConfig is used.
func NewObserveChecker(inner Checker, cfg ObserveConfig, logger *slog.Logger) Checker {
	if cfg.OnResult == nil {
		cfg = DefaultObserveConfig()
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &observeChecker{inner: inner, cfg: cfg, logger: logger}
}

func (o *observeChecker) Check(ctx context.Context, service string) Result {
	start := time.Now()
	r := o.inner.Check(ctx, service)
	elapsed := time.Since(start)
	o.logger.Debug("health check observed",
		"service", service,
		"status", r.Status.String(),
		"elapsed_ms", elapsed.Milliseconds(),
	)
	o.cfg.OnResult(service, r, elapsed)
	return r
}
