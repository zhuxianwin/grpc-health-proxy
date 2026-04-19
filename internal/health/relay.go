package health

import (
	"context"
	"log/slog"
	"os"
)

// RelayConfig controls how results are forwarded to a secondary sink.
type RelayConfig struct {
	// Sink receives a copy of every result. Must not be nil.
	Sink    func(service string, r Result)
	Logger  *slog.Logger
}

// DefaultRelayConfig returns a RelayConfig with a no-op sink.
func DefaultRelayConfig() RelayConfig {
	return RelayConfig{
		Sink:   func(_ string, _ Result) {},
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
}

type relayChecker struct {
	inner  Checker
	config RelayConfig
}

// NewRelayChecker wraps inner and forwards every result to config.Sink.
// The primary result is always returned unchanged.
func NewRelayChecker(inner Checker, cfg RelayConfig) Checker {
	if cfg.Sink == nil {
		cfg.Sink = func(_ string, _ Result) {}
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	return &relayChecker{inner: inner, config: cfg}
}

func (rc *relayChecker) Check(ctx context.Context, service string) Result {
	r := rc.inner.Check(ctx, service)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				rc.config.Logger.Warn("relay sink panicked", "panic", p)
			}
		}()
		rc.config.Sink(service, r)
	}()
	return r
}
