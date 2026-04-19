package health

import (
	"context"
	"log/slog"
)

// FilterConfig controls which results are forwarded to the inner checker.
type FilterConfig struct {
	// AllowedServices is a set of service names to allow. Empty means allow all.
	AllowedServices map[string]struct{}
	Logger          *slog.Logger
}

// DefaultFilterConfig returns a FilterConfig that allows all services.
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		AllowedServices: map[string]struct{}{},
		Logger:          slog.Default(),
	}
}

// FilterChecker wraps a Checker and only forwards calls for allowed services.
type FilterChecker struct {
	inner  Checker
	cfg    FilterConfig
	logger *slog.Logger
}

// NewFilterChecker returns a Checker that blocks services not in the allow list.
// If AllowedServices is empty, all services are forwarded.
func NewFilterChecker(inner Checker, cfg FilterConfig) *FilterChecker {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &FilterChecker{inner: inner, cfg: cfg, logger: cfg.Logger}
}

func (f *FilterChecker) Check(ctx context.Context, service string) Result {
	if len(f.cfg.AllowedServices) > 0 {
		if _, ok := f.cfg.AllowedServices[service]; !ok {
			f.logger.Debug("filter: service blocked", "service", service)
			return Result{Status: StatusUnhealthy, Err: ErrAdmissionLimitReached}
		}
	}
	return f.inner.Check(ctx, service)
}

// AllowedServices returns the current allow list.
func (f *FilterChecker) AllowedServices() map[string]struct{} {
	return f.cfg.AllowedServices
}
