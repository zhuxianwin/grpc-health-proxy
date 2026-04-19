package health

import (
	"context"
	"log/slog"
	"time"
)

// VersionConfig controls version-aware health checking.
type VersionConfig struct {
	// MinVersion is the minimum acceptable service version string.
	MinVersion string
	// Timeout for each check.
	Timeout time.Duration
}

// DefaultVersionConfig returns sensible defaults.
func DefaultVersionConfig() VersionConfig {
	return VersionConfig{
		Timeout: 5 * time.Second,
	}
}

// VersionChecker wraps a Checker and attaches the reported service version
// to the result metadata. If MinVersion is set and the reported version is
// lexicographically less than MinVersion the result is marked unhealthy.
type VersionChecker struct {
	inner   Checker
	cfg     VersionConfig
	logger  *slog.Logger
}

// NewVersionChecker creates a VersionChecker.
func NewVersionChecker(inner Checker, cfg VersionConfig, logger *slog.Logger) *VersionChecker {
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultVersionConfig().Timeout
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &VersionChecker{inner: inner, cfg: cfg, logger: logger}
}

// Check delegates to the inner checker and enriches the result with version metadata.
func (v *VersionChecker) Check(ctx context.Context, service string) Result {
	ctx, cancel := context.WithTimeout(ctx, v.cfg.Timeout)
	defer cancel()

	r := v.inner.Check(ctx, service)

	if r.Metadata == nil {
		r.Metadata = make(map[string]string)
	}

	reported, ok := r.Metadata["version"]
	if ok && v.cfg.MinVersion != "" && reported < v.cfg.MinVersion {
		v.logger.Warn("service version below minimum",
			"service", service,
			"version", reported,
			"min_version", v.cfg.MinVersion,
		)
		r.Status = StatusUnhealthy
		r.Err = ErrVersionTooLow
	}

	return r
}

// ErrVersionTooLow is returned when the reported version is below MinVersion.
var ErrVersionTooLow = versionError("service version below minimum required")

type versionError string

func (e versionError) Error() string { return string(e) }
