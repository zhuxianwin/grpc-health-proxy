package health

import (
	"context"
	"log/slog"
	"time"
)

// TraceConfig holds configuration for the trace checker.
type TraceConfig struct {
	// Logger is used to emit per-check trace lines.
	Logger *slog.Logger
}

// DefaultTraceConfig returns a TraceConfig with sensible defaults.
func DefaultTraceConfig(logger *slog.Logger) TraceConfig {
	if logger == nil {
		logger = slog.Default()
	}
	return TraceConfig{Logger: logger}
}

// traceChecker wraps a Checker and emits a structured log line for every check.
type traceChecker struct {
	inner  Checker
	cfg    TraceConfig
}

// NewTraceChecker returns a Checker that logs each call with its outcome and
// elapsed duration before returning the inner result unchanged.
func NewTraceChecker(inner Checker, cfg TraceConfig) Checker {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &traceChecker{inner: inner, cfg: cfg}
}

func (t *traceChecker) Check(ctx context.Context, service string) Result {
	start := time.Now()
	res := t.inner.Check(ctx, service)
	elapsed := time.Since(start)

	args := []any{
		"service", service,
		"status", res.Status.String(),
		"elapsed_ms", elapsed.Milliseconds(),
	}
	if res.Err != nil {
		args = append(args, "error", res.Err.Error())
	}
	t.cfg.Logger.DebugContext(ctx, "health check trace", args...)
	return res
}
