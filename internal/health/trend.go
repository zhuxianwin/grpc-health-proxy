package health

import (
	"context"
	"sync"
	"time"
)

// DefaultTrendConfig returns a TrendConfig with sensible defaults.
func DefaultTrendConfig() TrendConfig {
	return TrendConfig{
		Window:    10,
		MinSample: 3,
	}
}

// TrendConfig controls the behaviour of the trend checker.
type TrendConfig struct {
	// Window is the number of recent results to consider.
	Window int
	// MinSample is the minimum number of results required before a trend is
	// reported. Fewer results return the raw delegate result.
	MinSample int
}

// trendEntry records a single historical outcome.
type trendEntry struct {
	at      time.Time
	healthy bool
}

// TrendChecker wraps a Checker and enriches the Result with a "trend" metadata
// field: "improving", "degrading", or "stable".
type TrendChecker struct {
	inner  Checker
	cfg    TrendConfig
	mu     sync.Mutex
	history map[string][]trendEntry
}

// NewTrendChecker creates a TrendChecker that analyses the rolling window of
// outcomes for each service and annotates results with a trend label.
func NewTrendChecker(inner Checker, cfg TrendConfig) *TrendChecker {
	if cfg.Window <= 0 {
		cfg.Window = DefaultTrendConfig().Window
	}
	if cfg.MinSample <= 0 {
		cfg.MinSample = DefaultTrendConfig().MinSample
	}
	return &TrendChecker{
		inner:   inner,
		cfg:     cfg,
		history: make(map[string][]trendEntry),
	}
}

// Check delegates to the inner checker, records the outcome, and annotates the
// result with a trend label derived from the rolling window.
func (t *TrendChecker) Check(ctx context.Context, service string) Result {
	res := t.inner.Check(ctx, service)

	t.mu.Lock()
	entries := append(t.history[service], trendEntry{at: time.Now(), healthy: res.Status == StatusHealthy})
	if len(entries) > t.cfg.Window {
		entries = entries[len(entries)-t.cfg.Window:]
	}
	t.history[service] = entries
	trend := t.computeTrend(entries)
	t.mu.Unlock()

	if res.Meta == nil {
		res.Meta = make(map[string]string)
	}
	res.Meta["trend"] = trend
	return res
}

// computeTrend compares the first and second halves of the window to decide
// whether health is improving, degrading, or stable.
func (t *TrendChecker) computeTrend(entries []trendEntry) string {
	if len(entries) < t.cfg.MinSample {
		return "unknown"
	}
	mid := len(entries) / 2
	first := healthyRate(entries[:mid])
	second := healthyRate(entries[mid:])
	switch {
	case second > first+0.1:
		return "improving"
	case second < first-0.1:
		return "degrading"
	default:
		return "stable"
	}
}

func healthyRate(entries []trendEntry) float64 {
	if len(entries) == 0 {
		return 0
	}
	var healthy int
	for _, e := range entries {
		if e.healthy {
			healthy++
		}
	}
	return float64(healthy) / float64(len(entries))
}
