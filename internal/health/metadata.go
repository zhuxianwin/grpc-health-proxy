package health

import (
	"context"
	"log/slog"
	"maps"
)

// MetadataConfig holds configuration for the metadata checker.
type MetadataConfig struct {
	// Meta is the static key/value metadata to attach to every result.
	Meta map[string]string
}

// DefaultMetadataConfig returns a MetadataConfig with no metadata.
func DefaultMetadataConfig() MetadataConfig {
	return MetadataConfig{
		Meta: map[string]string{},
	}
}

// metadataChecker wraps a Checker and attaches static metadata to every result.
type metadataChecker struct {
	inner  Checker
	cfg    MetadataConfig
	logger *slog.Logger
}

// NewMetadataChecker returns a Checker that delegates to inner and merges cfg.Meta
// into the result's metadata. Existing metadata on the result takes precedence.
func NewMetadataChecker(inner Checker, cfg MetadataConfig, logger *slog.Logger) Checker {
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.Meta == nil {
		cfg.Meta = map[string]string{}
	}
	return &metadataChecker{inner: inner, cfg: cfg, logger: logger}
}

func (m *metadataChecker) Check(ctx context.Context, service string) Result {
	res := m.inner.Check(ctx, service)

	merged := make(map[string]string, len(m.cfg.Meta)+len(res.Metadata))
	// Static metadata first (lower priority).
	maps.Copy(merged, m.cfg.Meta)
	// Result metadata overwrites static keys.
	maps.Copy(merged, res.Metadata)

	if len(merged) > 0 {
		res.Metadata = merged
		m.logger.Debug("metadata attached", "service", service, "keys", len(merged))
	}
	return res
}
