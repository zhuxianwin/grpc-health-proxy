package health

import (
	"context"
	"log/slog"
)

// FallbackChecker wraps a primary Checker and falls back to a secondary
// Checker when the primary returns an error. Unhealthy (non-error) results
// from the primary are returned as-is without consulting the fallback.
type FallbackChecker struct {
	primary   Checker
	fallback  Checker
	logger    *slog.Logger
}

// NewFallbackChecker returns a FallbackChecker that tries primary first and
// delegates to fallback only on error.
func NewFallbackChecker(primary, fallback Checker, logger *slog.Logger) *FallbackChecker {
	if logger == nil {
		logger = slog.Default()
	}
	return &FallbackChecker{
		primary:  primary,
		fallback: fallback,
		logger:   logger,
	}
}

// Check calls the primary checker. If the primary returns an error the
// fallback is consulted and its result is returned instead.
func (f *FallbackChecker) Check(ctx context.Context, service string) Result {
	result := f.primary.Check(ctx, service)
	if result.Err == nil {
		return result
	}

	f.logger.Warn("primary checker failed, using fallback",
		"service", service,
		"error", result.Err,
	)

	fbResult := f.fallback.Check(ctx, service)
	fbResult.Source = "fallback"
	return fbResult
}
