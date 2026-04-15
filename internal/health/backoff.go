package health

import (
	"math"
	"time"
)

// BackoffConfig controls exponential backoff behaviour used by RetryChecker.
type BackoffConfig struct {
	// InitialInterval is the wait time before the first retry.
	InitialInterval time.Duration
	// Multiplier is applied to the interval after each attempt.
	Multiplier float64
	// MaxInterval caps the computed interval.
	MaxInterval time.Duration
	// Jitter adds up to this fraction of the current interval as random noise
	// (0 disables jitter).
	Jitter float64
}

// DefaultBackoffConfig returns sensible defaults for production use.
func DefaultBackoffConfig() BackoffConfig {
	return BackoffConfig{
		InitialInterval: 100 * time.Millisecond,
		Multiplier:      2.0,
		MaxInterval:     5 * time.Second,
		Jitter:          0.1,
	}
}

// Backoff computes the wait duration for the given attempt (0-indexed).
// It applies exponential growth capped at MaxInterval, then adds optional
// jitter using the provided rand source (0–1).
func (c BackoffConfig) Backoff(attempt int, randFrac float64) time.Duration {
	if c.InitialInterval <= 0 {
		return 0
	}

	growth := math.Pow(c.Multiplier, float64(attempt))
	interval := float64(c.InitialInterval) * growth

	if c.MaxInterval > 0 && interval > float64(c.MaxInterval) {
		interval = float64(c.MaxInterval)
	}

	if c.Jitter > 0 {
		interval += interval * c.Jitter * randFrac
	}

	return time.Duration(interval)
}
