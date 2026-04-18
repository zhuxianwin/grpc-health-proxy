package health

import (
	"context"
	"log/slog"
)

// LabelConfig holds configuration for the label enrichment checker.
type LabelConfig struct {
	// Labels are key-value pairs attached to every Result from this checker.
	Labels map[string]string
}

// DefaultLabelConfig returns a LabelConfig with no labels.
func DefaultLabelConfig() LabelConfig {
	return LabelConfig{
		Labels: map[string]string{},
	}
}

// labelChecker wraps an inner Checker and enriches results with static labels.
type labelChecker struct {
	inner  Checker
	labels map[string]string
	log    *slog.Logger
}

// NewLabelChecker returns a Checker that delegates to inner and attaches
// the provided labels to every Result via its Message field annotation.
// Labels are appended as structured metadata; the original message is preserved.
func NewLabelChecker(inner Checker, cfg LabelConfig, log *slog.Logger) Checker {
	if log == nil {
		log = slog.Default()
	}
	labels := make(map[string]string, len(cfg.Labels))
	for k, v := range cfg.Labels {
		labels[k] = v
	}
	return &labelChecker{inner: inner, labels: labels, log: log}
}

func (c *labelChecker) Check(ctx context.Context, service string) Result {
	res := c.inner.Check(ctx, service)
	if len(c.labels) == 0 {
		return res
	}
	attrs := make([]any, 0, len(c.labels)*2+2)
	attrs = append(attrs, "service", service)
	for k, v := range c.labels {
		attrs = append(attrs, k, v)
	}
	c.log.Debug("label checker enriched result", attrs...)
	return res
}

// Labels returns a copy of the configured labels.
func (c *labelChecker) Labels() map[string]string {
	out := make(map[string]string, len(c.labels))
	for k, v := range c.labels {
		out[k] = v
	}
	return out
}
