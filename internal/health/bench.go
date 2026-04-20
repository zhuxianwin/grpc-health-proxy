package health

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DefaultBenchConfig returns a BenchConfig with sensible defaults.
func DefaultBenchConfig() BenchConfig {
	return BenchConfig{
		WindowSize: 100,
		Logger:     slog.Default(),
	}
}

// BenchConfig controls the behaviour of the BenchChecker.
type BenchConfig struct {
	// WindowSize is the number of recent check durations to retain.
	WindowSize int
	// Logger receives structured diagnostic output.
	Logger *slog.Logger
}

// BenchChecker wraps a Checker and records per-check latency statistics.
type BenchChecker struct {
	inner  Checker
	cfg    BenchConfig
	mu     sync.Mutex
	window []time.Duration
}

// NewBenchChecker wraps inner and begins collecting latency samples.
func NewBenchChecker(inner Checker, cfg BenchConfig) *BenchChecker {
	if cfg.WindowSize <= 0 {
		cfg.WindowSize = DefaultBenchConfig().WindowSize
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &BenchChecker{inner: inner, cfg: cfg}
}

// Check delegates to the inner Checker, recording the elapsed duration.
func (b *BenchChecker) Check(ctx context.Context, service string) Result {
	start := time.Now()
	r := b.inner.Check(ctx, service)
	elapsed := time.Since(start)

	b.mu.Lock()
	b.window = append(b.window, elapsed)
	if len(b.window) > b.cfg.WindowSize {
		b.window = b.window[len(b.window)-b.cfg.WindowSize:]
	}
	b.mu.Unlock()

	b.cfg.Logger.Debug("bench check", "service", service,
		"elapsed_ms", elapsed.Milliseconds(), "status", r.Status)
	return r
}

// Stats returns the mean and maximum observed duration across the current window.
func (b *BenchChecker) Stats() (mean, max time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.window) == 0 {
		return 0, 0
	}
	var total time.Duration
	for _, d := range b.window {
		total += d
		if d > max {
			max = d
		}
	}
	return total / time.Duration(len(b.window)), max
}
