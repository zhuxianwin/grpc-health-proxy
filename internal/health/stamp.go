package health

import (
	"context"
	"log/slog"
	"time"
)

// StampConfig controls how timestamps are attached to health results.
type StampConfig struct {
	// TimeFunc overrides time.Now for testing.
	TimeFunc func() time.Time
}

// DefaultStampConfig returns a StampConfig with sensible defaults.
func DefaultStampConfig() StampConfig {
	return StampConfig{TimeFunc: time.Now}
}

// stampChecker wraps a Checker and stamps each result with a checked_at timestamp.
type stampChecker struct {
	inner  Checker
	cfg    StampConfig
	logger *slog.Logger
}

// NewStampChecker returns a Checker that attaches a "checked_at" metadata field
// (RFC3339) to every result returned by inner.
func NewStampChecker(inner Checker, cfg StampConfig, logger *slog.Logger) Checker {
	if cfg.TimeFunc == nil {
		cfg.TimeFunc = time.Now
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &stampChecker{inner: inner, cfg: cfg, logger: logger}
}

func (s *stampChecker) Check(ctx context.Context, service string) Result {
	res := s.inner.Check(ctx, service)
	if res.Metadata == nil {
		res.Metadata = make(map[string]string)
	}
	res.Metadata["checked_at"] = s.cfg.TimeFunc().UTC().Format(time.RFC3339)
	return res
}
