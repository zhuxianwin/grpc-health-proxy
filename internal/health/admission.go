package health

import (
	"context"
	"log/slog"
	"sync/atomic"
)

// AdmissionConfig controls the admission checker behaviour.
type AdmissionConfig struct {
	// MaxConcurrent is the maximum number of concurrent health checks allowed.
	// Checks beyond this limit are rejected with an error.
	MaxConcurrent int64
}

// DefaultAdmissionConfig returns a sensible default configuration.
func DefaultAdmissionConfig() AdmissionConfig {
	return AdmissionConfig{
		MaxConcurrent: 10,
	}
}

// admissionChecker wraps a Checker and enforces a concurrency limit using an
// atomic counter rather than a channel so the hot path is allocation-free.
type admissionChecker struct {
	inner   Checker
	max     int64
	active  atomic.Int64
	logger  *slog.Logger
}

// NewAdmissionChecker returns a Checker that rejects requests when more than
// cfg.MaxConcurrent checks are already in flight.
func NewAdmissionChecker(inner Checker, cfg AdmissionConfig, logger *slog.Logger) Checker {
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = DefaultAdmissionConfig().MaxConcurrent
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &admissionChecker{
		inner:  inner,
		max:    cfg.MaxConcurrent,
		logger: logger,
	}
}

func (a *admissionChecker) Check(ctx context.Context, service string) Result {
	current := a.active.Add(1)
	defer a.active.Add(-1)

	if current > a.max {
		a.logger.Warn("admission limit reached",
			"service", service,
			"active", current,
			"max", a.max,
		)
		return Result{
			Status: StatusUnknown,
			Err:    ErrAdmissionLimitExceeded,
		}
	}

	return a.inner.Check(ctx, service)
}
