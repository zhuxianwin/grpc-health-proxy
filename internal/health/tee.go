package health

import (
	"context"
	"log/slog"
	"os"
)

// TeeConfig holds configuration for the tee checker.
type TeeConfig struct {
	// Logger is used for diagnostic output. Defaults to slog.Default() if nil.
	Logger *slog.Logger
}

// DefaultTeeConfig returns a TeeConfig with sensible defaults.
func DefaultTeeConfig() TeeConfig {
	return TeeConfig{
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
}

// teeChecker delegates to a primary checker and forwards every result to one
// or more sink checkers for side-effect purposes (e.g. recording, mirroring,
// or feeding secondary pipelines). Unlike MirrorChecker, sinks receive the
// same service name and their results are discarded — the primary result is
// always returned to the caller.
type teeChecker struct {
	inner  Checker
	sinks  []Checker
	logger *slog.Logger
}

// NewTeeChecker returns a Checker that calls inner, then fans the result out
// to each sink in order. Sink errors are logged but never propagated.
func NewTeeChecker(inner Checker, sinks []Checker, cfg TeeConfig) Checker {
	if cfg.Logger == nil {
		cfg.Logger = DefaultTeeConfig().Logger
	}
	return &teeChecker{
		inner:  inner,
		sinks:  sinks,
		logger: cfg.Logger,
	}
}

// Check calls the inner checker, then delivers the result to every sink.
// Sinks are called sequentially in the order they were registered. A panic or
// error from a sink is recovered and logged so that the caller always receives
// the primary result.
func (t *teeChecker) Check(ctx context.Context, service string) Result {
	result := t.inner.Check(ctx, service)

	for i, sink := range t.sinks {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.logger.Warn("tee sink panicked",
						"sink_index", i,
						"service", service,
						"panic", r,
					)
				}
			}()
			sinkResult := sink.Check(ctx, service)
			if sinkResult.Err != nil {
				t.logger.Debug("tee sink returned error",
					"sink_index", i,
					"service", service,
					"error", sinkResult.Err,
				)
			}
		}()
	}

	return result
}
