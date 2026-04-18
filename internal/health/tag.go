package health

import (
	"context"
	"log/slog"
)

// TagConfig holds configuration for the tag checker.
type TagConfig struct {
	Tags map[string]string
}

// DefaultTagConfig returns a TagConfig with no tags.
func DefaultTagConfig() TagConfig {
	return TagConfig{Tags: map[string]string{}}
}

// tagChecker wraps a Checker and attaches static metadata tags to each Result.
type tagChecker struct {
	inner  Checker
	tags   map[string]string
	logger *slog.Logger
}

// NewTagChecker returns a Checker that delegates to inner and attaches tags to
// every Result it returns.
func NewTagChecker(inner Checker, cfg TagConfig, logger *slog.Logger) Checker {
	if logger == nil {
		logger = slog.Default()
	}
	tags := make(map[string]string, len(cfg.Tags))
	for k, v := range cfg.Tags {
		tags[k] = v
	}
	return &tagChecker{inner: inner, tags: tags, logger: logger}
}

func (t *tagChecker) Check(ctx context.Context, service string) Result {
	res := t.inner.Check(ctx, service)
	if len(t.tags) == 0 {
		return res
	}
	merged := make(map[string]string, len(res.Tags)+len(t.tags))
	for k, v := range res.Tags {
		merged[k] = v
	}
	for k, v := range t.tags {
		merged[k] = v
	}
	res.Tags = merged
	return res
}

// Tags returns the static tags attached by this checker.
func (t *tagChecker) Tags() map[string]string {
	out := make(map[string]string, len(t.tags))
	for k, v := range t.tags {
		out[k] = v
	}
	return out
}
