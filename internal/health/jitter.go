package health

import (
	"context"
	"log/slog"
	"math/rand"
	"time"
)

// JitterConfig controls how much random delay is added before each check.
type JitterConfig struct {
	// MaxJitter is the upper bound of the random delay added before delegating.
	MaxJitter time.Duration
}

// DefaultJitterConfig returns a sensible default.
func DefaultJitterConfig() JitterConfig {
	return JitterConfig{
		MaxJitter: 50 * time.Millisecond,
	}
}

// jitterChecker wraps a Checker and adds a random pre-check delay.
type jitterChecker struct {
	inner  Checker
	cfg    JitterConfig
	logger *slog.Logger
}

// NewJitterChecker returns a Checker that sleeps a random duration in
// [0, cfg.MaxJitter) before delegating to inner.
func NewJitterChecker(inner Checker, cfg JitterConfig, logger *slog.Logger) Checker {
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.MaxJitter <= 0 {
		cfg = DefaultJitterConfig()
	}
	return &jitterChecker{inner: inner, cfg: cfg, logger: logger}
}

func (j *jitterChecker) Check(ctx context.Context, service string) Result {
	delay := time.Duration(rand.Int63n(int64(j.cfg.MaxJitter)))
	j.logger.Debug("jitter delay", "service", service, "delay", delay)
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return Result{Status: StatusUnhealthy, Err: ctx.Err()}
	}
	return j.inner.Check(ctx, service)
}
