package health

import (
	"sync"
	"time"
)

// WindowConfig controls the sliding window checker behaviour.
type WindowConfig struct {
	// Size is the number of recent results to track.
	Size int
	// Threshold is the minimum fraction [0,1] of healthy results required.
	Threshold float64
}

// DefaultWindowConfig returns sensible defaults.
func DefaultWindowConfig() WindowConfig {
	return WindowConfig{Size: 10, Threshold: 0.6}
}

// WindowChecker wraps a Checker and only reports healthy when the fraction of
// healthy results in a sliding window meets the configured threshold.
type WindowChecker struct {
	inner  Checker
	cfg    WindowConfig
	mu     sync.Mutex
	bucket []bool
}

// NewWindowChecker returns a WindowChecker. Zero-value cfg fields are replaced
// with defaults.
func NewWindowChecker(inner Checker, cfg WindowConfig) *WindowChecker {
	def := DefaultWindowConfig()
	if cfg.Size <= 0 {
		cfg.Size = def.Size
	}
	if cfg.Threshold <= 0 {
		cfg.Threshold = def.Threshold
	}
	return &WindowChecker{inner: inner, cfg: cfg}
}

// Check delegates to the inner checker, records the result in the sliding
// window, and overrides a healthy result with unhealthy when the window
// fraction falls below the threshold.
func (w *WindowChecker) Check(ctx interface{ Deadline() (time.Time, bool) }, service string) Result {
	// ctx is context.Context; use the health.Checker interface signature.
	type checker interface {
		Check(ctx interface{ Deadline() (time.Time, bool) }, service string) Result
	}
	res := w.inner.(checker).Check(ctx, service)

	w.mu.Lock()
	w.bucket = append(w.bucket, res.Status == StatusHealthy)
	if len(w.bucket) > w.cfg.Size {
		w.bucket = w.bucket[len(w.bucket)-w.cfg.Size:]
	}
	healthy := 0
	for _, v := range w.bucket {
		if v {
			healthy++
		}
	}
	fraction := float64(healthy) / float64(len(w.bucket))
	w.mu.Unlock()

	if res.Status == StatusHealthy && fraction < w.cfg.Threshold {
		return Result{Service: service, Status: StatusUnhealthy, Err: nil}
	}
	return res
}

// WindowStats returns the current window size and healthy fraction for
// observability purposes.
func (w *WindowChecker) WindowStats() (size int, fraction float64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.bucket) == 0 {
		return 0, 0
	}
	healthy := 0
	for _, v := range w.bucket {
		if v {
			healthy++
		}
	}
	return len(w.bucket), float64(healthy) / float64(len(w.bucket))
}
