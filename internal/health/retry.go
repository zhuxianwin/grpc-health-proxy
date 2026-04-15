package health

import (
	"context"
	"time"
)

// RetryConfig holds configuration for retry behaviour on health check failures.
type RetryConfig struct {
	// Attempts is the total number of attempts (including the first).
	Attempts int
	// Delay is the wait time between consecutive attempts.
	Delay time.Duration
}

// DefaultRetryConfig returns a sensible default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		Attempts: 3,
		Delay:    200 * time.Millisecond,
	}
}

// RetryChecker wraps a Checker and retries transient failures according to
// the provided RetryConfig. Only errors (not explicit UNHEALTHY status) are
// retried so that a service that intentionally reports itself unhealthy is
// respected immediately.
type RetryChecker struct {
	inner  Checker
	cfg    RetryConfig
}

// Checker is the interface satisfied by health checkers.
type Checker interface {
	Check(ctx context.Context, service string) Result
}

// NewRetryChecker wraps inner with retry logic using cfg.
func NewRetryChecker(inner Checker, cfg RetryConfig) *RetryChecker {
	if cfg.Attempts <= 0 {
		cfg.Attempts = 1
	}
	return &RetryChecker{inner: inner, cfg: cfg}
}

// Check executes the underlying checker, retrying on error up to cfg.Attempts
// times. If the final attempt still returns an error the last Result is
// returned unchanged.
func (r *RetryChecker) Check(ctx context.Context, service string) Result {
	var result Result
	for attempt := 0; attempt < r.cfg.Attempts; attempt++ {
		result = r.inner.Check(ctx, service)
		// Do not retry if the service explicitly responded (healthy or not).
		if result.Err == nil {
			return result
		}
		// Last attempt — return whatever we got.
		if attempt == r.cfg.Attempts-1 {
			break
		}
		select {
		case <-ctx.Done():
			return Result{Status: StatusUnknown, Err: ctx.Err()}
		case <-time.After(r.cfg.Delay):
		}
	}
	return result
}
