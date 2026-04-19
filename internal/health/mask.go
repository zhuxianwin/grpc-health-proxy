package health

import (
	"context"
	"log/slog"
	"os"
)

// MaskConfig controls which metadata keys are redacted from results.
type MaskConfig struct {
	// Keys lists metadata keys whose values will be replaced with "***".
	Keys []string
	// Logger is used for debug output; defaults to slog.Default().
	Logger *slog.Logger
}

// DefaultMaskConfig returns a MaskConfig with no masked keys.
func DefaultMaskConfig() MaskConfig {
	return MaskConfig{
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
}

type maskChecker struct {
	inner  Checker
	keys   map[string]struct{}
	logger *slog.Logger
}

// NewMaskChecker wraps inner and redacts the listed metadata keys from every
// Result before returning it to the caller.
func NewMaskChecker(inner Checker, cfg MaskConfig) Checker {
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	set := make(map[string]struct{}, len(cfg.Keys))
	for _, k := range cfg.Keys {
		set[k] = struct{}{}
	}
	return &maskChecker{inner: inner, keys: set, logger: cfg.Logger}
}

func (m *maskChecker) Check(ctx context.Context, service string) Result {
	res := m.inner.Check(ctx, service)
	if len(m.keys) == 0 || len(res.Metadata) == 0 {
		return res
	}
	masked := make(map[string]string, len(res.Metadata))
	for k, v := range res.Metadata {
		if _, redact := m.keys[k]; redact {
			masked[k] = "***"
			m.logger.Debug("masked metadata key", "key", k)
		} else {
			masked[k] = v
		}
	}
	res.Metadata = masked
	return res
}
