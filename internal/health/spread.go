package health

import (
	"context"
	"log/slog"
	"math/rand"
	"os"
	"time"
)

// DefaultSpreadConfig returns a SpreadConfig with sensible defaults.
func DefaultSpreadConfig() SpreadConfig {
	return SpreadConfig{
		MaxDelay: 500 * time.Millisecond,
	}
}

// SpreadConfig controls the spread checker behaviour.
type SpreadConfig struct {
	// MaxDelay is the upper bound of the random delay injected before
	// delegating to the inner checker. Zero uses the default.
	MaxDelay time.Duration
}

// spreadChecker wraps a Checker and introduces a random delay before each
// check to spread load across replicas that would otherwise fire in lockstep.
type spreadChecker struct {
	inner  Checker
	cfg    SpreadConfig
	logger *slog.Logger
}

// NewSpreadChecker returns a Checker that sleeps for a random duration in
// [0, cfg.MaxDelay) before delegating to inner. This is useful when many
// sidecar replicas start simultaneously and would otherwise hammer the
// upstream gRPC service at the same instant.
func NewSpreadChecker(inner Checker, cfg SpreadConfig, logger *slog.Logger) Checker {
	if cfg.MaxDelay <= 0 {
		cfg = DefaultSpreadConfig()
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	return &spreadChecker{inner: inner, cfg: cfg, logger: logger}
}

func (s *spreadChecker) Check(ctx context.Context, service string) Result {
	delay := time.Duration(rand.Int63n(int64(s.cfg.MaxDelay)))
	s.logger.Debug("spread: injecting delay", "service", service, "delay", delay)

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return Result{Status: StatusUnhealthy, Err: ctx.Err()}
	}

	return s.inner.Check(ctx, service)
}
