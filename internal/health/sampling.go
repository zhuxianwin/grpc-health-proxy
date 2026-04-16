package health

import (
	"context"
	"log/slog"
	"math/rand"
	"sync/atomic"
)

// SamplingConfig controls what fraction of health checks are forwarded.
type SamplingConfig struct {
	// Rate is the probability [0.0, 1.0] that a check is executed.
	// Checks that are not sampled return the last known result, or Healthy if none.
	Rate float64
}

// DefaultSamplingConfig returns a config that samples every check.
func DefaultSamplingConfig() SamplingConfig {
	return SamplingConfig{Rate: 1.0}
}

// SamplingChecker wraps a Checker and probabilistically skips checks.
type SamplingChecker struct {
	inner  Checker
	rate   float64
	last   atomic.Pointer[Result]
	logger *slog.Logger
}

// NewSamplingChecker returns a Checker that forwards calls with probability cfg.Rate.
func NewSamplingChecker(inner Checker, cfg SamplingConfig, logger *slog.Logger) *SamplingChecker {
	rate := cfg.Rate
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &SamplingChecker{inner: inner, rate: rate, logger: logger}
}

// Check runs the underlying check with probability s.rate.
// On a skip it returns the last cached result (Healthy if never checked).
func (s *SamplingChecker) Check(ctx context.Context, service string) Result {
	if s.rate < 1.0 && rand.Float64() > s.rate {
		if prev := s.last.Load(); prev != nil {
			s.logger.Debug("sampling: skipped check, returning cached result",
				"service", service, "status", prev.Status)
			return *prev
		}
		s.logger.Debug("sampling: skipped check, no cache — assuming healthy", "service", service)
		return Result{Status: StatusHealthy}
	}
	r := s.inner.Check(ctx, service)
	s.last.Store(&r)
	return r
}
